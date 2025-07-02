package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Step 1: Connect to 'postgres' database as root
	adminConn := "postgres://postgres:root@sms-db:5432/postgres?sslmode=disable"
	adminDB, err := sql.Open("postgres", adminConn)
	if err != nil {
		log.Fatal("❌ Error connecting to admin DB:", err)
	}
	defer adminDB.Close()

	// Retry in case DB isn't ready yet
	for i := 0; i < 10; i++ {
		err = adminDB.Ping()
		if err == nil {
			break
		}
		fmt.Println("⏳ Waiting for postgres to be ready...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("❌ Database never became ready:", err)
	}

	fmt.Println("🔧 Creating admin role if not exists...")
	_, err = adminDB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'admin') THEN
			CREATE ROLE admin LOGIN PASSWORD 'admin';
		END IF;
	END $$;`)
	if err != nil {
		log.Fatal("❌ Error creating admin role:", err)
	}

	fmt.Println("📦 Checking if smsdb database exists...")
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'smsdb')").Scan(&exists)
	if err != nil {
		log.Fatal("❌ Error checking database existence:", err)
	}

	if !exists {
		fmt.Println("📦 Creating 'smsdb' database...")
		// CREATE DATABASE must be outside any transaction
		_, err = adminDB.Exec(`CREATE DATABASE smsdb OWNER admin`)
		if err != nil {
			log.Fatal("❌ Error creating smsdb database:", err)
		}
		fmt.Println("✅ 'smsdb' database created.")
	} else {
		fmt.Println("ℹ️ 'smsdb' database already exists.")
	}

	// Step 2: Connect to smsdb as admin
	userConn := "postgres://admin:admin@db:5432/smsdb?sslmode=disable"
	db, err := sql.Open("postgres", userConn)
	if err != nil {
		log.Fatal("❌ Error connecting to smsdb:", err)
	}
	defer db.Close()

	// Wait for it to come up
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		fmt.Println("⏳ Waiting for smsdb to be ready...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("❌ smsdb never became ready:", err)
	}

	fmt.Println("🗑️ Dropping all existing tables...")
	dropAllTables := `
	DO $$ 
	DECLARE
	r RECORD;
	BEGIN
	FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
	EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
	END LOOP;
	END $$;`
	_, err = db.ExecContext(context.Background(), dropAllTables)
	if err != nil {
		log.Fatal("❌ Dropping tables failed:", err)
	}

	// ✅ Create tables
	createStatements := []struct {
		name string
		sql  string
	}{
		{"users", `
			CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL UNIQUE,
				role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'viewer')),
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(100) NOT NULL
			);`},

		{"user_sessions", `
			CREATE TABLE IF NOT EXISTS user_sessions (
				id SERIAL PRIMARY KEY,
				username TEXT NOT NULL UNIQUE REFERENCES users(name) ON DELETE CASCADE,
				refresh_token TEXT NOT NULL UNIQUE,
				expires_at TIMESTAMPTZ NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now()
			);`},

		{"server_devices", `
			CREATE TABLE IF NOT EXISTS server_devices (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				ip VARCHAR(45) NOT NULL UNIQUE,
				tag VARCHAR(100) NOT NULL DEFAULT '',
				os VARCHAR(100) NOT NULL DEFAULT '',
				access_token VARCHAR(255) NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
			);`},

		{"alerts", `
			CREATE TABLE IF NOT EXISTS alerts (
				id SERIAL PRIMARY KEY,
				host VARCHAR(45) NOT NULL,
				severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
				content TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'notseen' CHECK (status IN ('notseen', 'seen')),
				time TIMESTAMPTZ DEFAULT now()
			);`},

		{"mac_access_status", `
			CREATE TABLE IF NOT EXISTS mac_access_status (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				mac VARCHAR(17) NOT NULL,
				status VARCHAR(20) NOT NULL CHECK (status IN ('BLACKLISTED', 'WHITELISTED')),
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP NOT NULL DEFAULT NOW()
			);`},
	}

	for _, stmt := range createStatements {
		fmt.Printf("📦 Creating table: %s...\n", stmt.name)
		if _, err := db.ExecContext(context.Background(), stmt.sql); err != nil {
			log.Fatalf("❌ Failed to create %s: %v", stmt.name, err)
		}
	}

	// 👤 Insert default admin
	fmt.Println("👤 Inserting default admin user...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("❌ Hashing password failed:", err)
	}

	insertAdmin := `
INSERT INTO users (name, role, email, password_hash)
SELECT * FROM (
  SELECT $1::text, $2::text, $3::text, $4::text
) AS tmp
WHERE NOT EXISTS (
  SELECT 1 FROM users WHERE email = $3
);`

	result, err := db.ExecContext(context.Background(), insertAdmin, "admin", "admin", "admin@example.com", string(hashedPassword))
	if err != nil {
		log.Fatal("❌ Inserting admin user failed:", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		fmt.Println("✅ Admin user created")
	} else {
		fmt.Println("ℹ️ Admin user already exists")
	}

	fmt.Println("\n🎉 Database initialized successfully!")
	fmt.Println("📊 Tables: users, user_sessions, server_devices, alerts, mac_access_status")
	fmt.Println("👤 Username: admin | Password: admin | Email: admin@example.com")
}
