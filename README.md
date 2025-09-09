# Server Management Suite (SMS)
A centralized, secure, and scalable platform for managing Linux and Windows servers across institutional networks, data centers, and enterprises.

---

## 🧭 Overview
**Server Management Suite (SMS)** is an all-in-one server management solution designed to simplify and centralize the administration of multiple Linux-based and Windows-based servers. Through a modern web interface, SMS allows system administrators to:
- Monitor server health in real time
- Configure system and network settings remotely
- Manage users and roles
- Receive proactive alerts
- Streamline operations across fleets of servers

Built with a modular architecture and robust security, SMS is ideal for universities, enterprises, and data centers seeking to reduce downtime and improve operational efficiency.

---

## 📦 Installation
The **Server Management Suite (SMS)** can be deployed using two primary methods:
1. **Docker-based Deployment** (Recommended for most users)
2. **Manual Setup** (Traditional DevOps installation for more control or debugging)

We will first explore the Docker method and then explain the manual approach.

## 🐳 Docker Setup (Recommended)
This is the easiest and fastest method to get SMS running on any machine.

### ✅ Prerequisites
**For Windows:**
* Git for Windows
* Docker Desktop
* Windows Subsystem for Linux (WSL2)
* Enable Virtualization and Hyper-V in BIOS & Windows Features

**For Linux:**
* `git`
* `docker` and `docker-compose` (`docker compose` v2 CLI)

### ⚙️ Setup Instructions
We provide platform-specific setup scripts:
* For **Linux**: Run `./setup.sh`
* For **Windows (PowerShell)**: Run `./setup.ps1`

To use the script:
```bash
./setup.sh start      # Start setup
./setup.sh clean      # Remove containers, volumes, images
./setup.sh help       # Show help menu
```

Equivalent commands apply for PowerShell on Windows.

The script:
* Lets you select a host IP
* Prompts for backend and DB ports
* Generates `.env` files for both frontend and backend
* Builds Docker images
* Initializes the PostgreSQL database
* Launches frontend/backend services

Once setup completes, your system is live.

## ⚙️ Manual Setup (For DevOps / Debugging)

### ✅ Prerequisites
* PostgreSQL (v15+)
* Go (v1.20+)
* sqlc
* Node.js (v18+)
* Git

### Setup Steps
```bash
git clone https://github.com/kishore-001/ServerManagementSuite.git
cd ServerManagementSuite
```

### 📂 Frontend Setup
```bash
cd frontend
npm install
```

Create `.env` file in `/frontend`:
```
VITE_BACKEND_URL=http://<your_backend_ip>:9000
```

Build the frontend:
```bash
npm run build
```

### 🔌 Backend Setup
```bash
cd backend
cd db
sqlc generate
cd ..
```

Create `.env` in `/backend`:
```
DATABASE_URL=postgres://<user>:<pass>@<host>:<port>/<db>?sslmode=disable
CLIENT_PORT=2210
CLIENT_PROTOCOL=http
JWT_SECRET=vanakamdamapla
SERVER_PORT=8000
LOG_LEVEL=info
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=servermanagementcit@gmail.com
SMTP_PASSWORD=tgekktudhggxwpok
SMTP_FROM=SMS Alerts <servermanagementcit@gmail.com>
```

Build backend:
```bash
go build -o server main.go
```

### 🗃️ PostgreSQL Setup
1. Install PostgreSQL and run:
```bash
sudo systemctl start postgresql  # for linux
```
or download PostgreSQL from the installer

Create user and database:
```sql
CREATE USER admin WITH PASSWORD 'admin';
CREATE DATABASE smsdb;
GRANT ALL PRIVILEGES ON DATABASE smsdb TO admin;
```

Grant permissions on public schema:
```sql
GRANT ALL ON SCHEMA public TO admin;
```

Confirm DB connection (CLI or Go):
```bash
psql -U admin -d smsdb -h localhost -p 9001
```

Sample DB URL:
```
postgres://admin:admin@sms-db:5432/smsdb?sslmode=disable
```

Initialize DB structure:
```bash
go run temp/dbinit.go
```

## 🖥️ Server Controller Setup (Formerly "Client")

The Server Controller runs on each target server and communicates with the backend.

### ✅ Prerequisites
* Go (1.20+)
* Git

### 🐧 For Linux Servers
```bash
git clone https://github.com/kishore-001/ServerManagementSuite.git
cd ServerManagementSuite
go mod tidy
go build -o controller main.go
./controller
```

### 🪟 For Windows Servers
```bash
git clone https://github.com/kishore-001/ServerManagementSuite.git
cd ServerManagementSuite
go mod tidy
go build -o controller.exe main.go
.\controller.exe
```

This runs on port 2210 by default.

## 🔐 Token Setup

On first run, the Server Controller will request a client token (generated from the frontend admin panel). If you want to reset or change the token:

* **Linux**: Delete `client/linux/auth/token.hash`
* **Windows**: Delete `client/windows/auth/token.hash`

After deletion, re-run the controller and it will ask for a new token.

---

## ⚙️ Working of the System
The **Server Management Suite (SMS)** operates as a centralized monitoring and management system for all registered servers and network devices. Once deployed, the backend service listens on the specified port (default: `9000`) and exposes API endpoints for the frontend and server controllers. The frontend interface provides an admin-only dashboard to monitor, configure, and control connected devices.

Each server (Windows/Linux) runs a lightweight **server controller agent** that connects to the central backend. These agents continuously report device health, system logs, CPU/memory/disk/network statistics, and also accept configuration or command updates from the backend. Routers and firewalls (like Fortinet) are managed via SSH or SNMP integrations from the backend.

The admin user can view health summaries, trigger backups, monitor real-time alerts, update configurations, and even restart specific services on the managed servers or network devices — all from a single dashboard. All communication between server controllers and the backend is authenticated using a secure token-based mechanism to prevent unauthorized access.

---

## 🏗️ Architecture Overview
The architecture of the Server Management Suite follows a **modular client-server model**. It has three main components:

1. **Frontend Web Interface:** Built using React + Vite, this is the admin dashboard where users can manage devices, view alerts, and configure system settings. It connects directly to the backend via REST API.

2. **Backend API Server:** A Go-based service that handles user authentication, API routing, database operations, and device coordination. It integrates with PostgreSQL for persistent data and provides endpoints for both the frontend and server controllers.

3. **Server Controllers (Agents):** Lightweight Go programs running on each server or device. These send periodic health reports, logs, and accept backend-issued commands. They are authenticated using tokens generated by the backend.

All components interact over HTTP, using ports defined during setup. The backend also manages token validation, alert generation, and logging of all activity.

### 🔄 Data Flow Summary:
1. Admin logs in to the **frontend** and registers devices.
2. The **server controller** running on each device connects to the backend and authenticates using its token.
3. The backend records health and performance metrics in the database.
4. The admin can issue commands or configuration changes via frontend.
5. The backend relays instructions securely to the correct agent.

### 🖼️ Architecture Diagram (To Be Added)
*[Architecture diagram will be provided]*

---

## ✅ Maintainers / Contact

```
Maintained by:
- Kishore [@kishore-001](https://github.com/kishore-001)
- Karthick Raghul [@karthickRaghul](https://github.com/KarthickRaghul)

```
