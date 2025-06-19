package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:8500/SSMS?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 🗑️ Drop all existing tables first
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
		log.Fatal("dropping existing tables:", err)
	}
	fmt.Println("✅ All existing tables dropped")

	// 📊 Create users table with UNIQUE constraint on name
	fmt.Println("📊 Creating users table...")
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL UNIQUE,
        role VARCHAR(50) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(100) NOT NULL
    );`

	_, err = db.ExecContext(context.Background(), createUsersTable)
	if err != nil {
		log.Fatal("creating users table:", err)
	}
	fmt.Println("✅ Users table created")

	// 🔐 Create user_sessions table
	fmt.Println("🔐 Creating user_sessions table...")
	createUserSessionsTable := `
    CREATE TABLE IF NOT EXISTS user_sessions (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL REFERENCES users(name) ON DELETE CASCADE,
        refresh_token TEXT NOT NULL UNIQUE,
        expires_at TIMESTAMPTZ NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );`

	_, err = db.ExecContext(context.Background(), createUserSessionsTable)
	if err != nil {
		log.Fatal("creating user_sessions table:", err)
	}
	fmt.Println("✅ User_sessions table created")

	// 📊 Create server devices table with UNIQUE constraint on name
	fmt.Println("📊 Creating server devices table...")
	createServerDevices := `
   CREATE TABLE IF NOT EXISTS server_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip VARCHAR(45) NOT NULL UNIQUE,
    tag VARCHAR(100) NOT NULL DEFAULT '',  -- NOT NULL with default
    os VARCHAR(100) NOT NULL DEFAULT '',   -- NOT NULL with default
    access_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

	_, err = db.ExecContext(context.Background(), createServerDevices)
	if err != nil {
		log.Fatal("creating server devices table:", err)
	}
	fmt.Println("✅ server Devices table created")

	createServerAlerts := `
	CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    host VARCHAR(45) NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    content TEXT NOT NULL,
    time TIMESTAMPTZ DEFAULT now()
  );
	`

	_, err = db.ExecContext(context.Background(), createServerAlerts)
	if err != nil {
		log.Fatal("creating server alerts table:", err)
	}
	fmt.Println("✅ server alerts table created")

	// 🔑 Hash default admin password
	fmt.Println("🔑 Creating default admin user...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("hashing password:", err)
	}

	// 👤 Insert default admin user
	insertUser := `
    INSERT INTO users (name, role, email, password_hash)
	SELECT $1::text, $2::text, $3::text, $4::text
    WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = $3);`

	result, err := db.ExecContext(context.Background(), insertUser, "admin", "admin", "admin@example.com", string(hashedPassword))
	if err != nil {
		log.Fatal("inserting user:", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Println("✅ Default admin user created")
	} else {
		fmt.Println("ℹ️ Admin user already exists")
	}

	// 🎉 Success summary
	fmt.Println("\n🎉 Database setup complete!")
	fmt.Println("📊 Tables: users, user_sessions")
	fmt.Println("👤 Default admin credentials:")
	fmt.Println("   Username: admin")
	fmt.Println("   Password: admin")
	fmt.Println("   Email: admin@example.com")
	fmt.Println("   Role: admin")
}
