# Server Management Suite (SMS)

A centralized, secure, and scalable platform for managing Linux servers across institutional networks, data centers, and enterprises.

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

## 🔑 Key Features

- **Centralized Dashboard:** Manage all servers from a unified portal—no more individual SSH sessions.
- **Real-time Monitoring:** Live metrics for CPU, memory, disk, and network I/O, updated every 30 seconds with historical graphs.
- **Configuration Management:** Remotely edit hostnames, network interfaces, firewall rules, and more.
- **Alert System:** Automated notifications for resource thresholds (CPU, RAM, disk, network), with severity levels.
- **Role-Based Access Control:** Admin (full access) and Viewer (read-only) roles, with UI tailored to each.
- **Secure Communication:** JWT authentication, access tokens, input validation, and encrypted protocols (HTTPS).
- **Modular Architecture:** Easily extensible for new features and scalable to large server fleets.

---

## 🛠️ Installation Guide

### Prerequisites

- Go 1.19+ (backend, client)
- PostgreSQL (database)
- Node.js & npm (frontend)
- Linux or Windows servers for client agents


### Backend Setup (Go + PostgreSQL)

```bash
git clone <repository-url>
cd backend
go mod download
cp .env.example .env # Edit with your DB credentials and secrets
go run temp/dbinit.go # Initialize database
go run main.go # Start backend server
```

### Frontend Setup (React + Vite)

```bash
cd frontend
npm install
cp .env.example .env # Set backend URL
npm run dev # Start frontend (dev mode)
npm run build # Build for production
```

### Client Deployment (Go Agent)

On each managed Linux server:

```bash
sudo ./client-executable --port 2210
```

On each managed Windows server:
use admin powershell or command prompt to run the exe file



- Registers with backend using access token
- Collects and sends metrics every 30 seconds

---

## 📐 System Architecture

| Component | Technology      | Role                                         |
|-----------|----------------|----------------------------------------------|
| Frontend  | React + Vite   | User dashboard, real-time monitoring         |
| Backend   | Go + PostgreSQL| API, logic, authentication, data storage     |
| Client    | Go             | Runs on each server, collects metrics, executes commands |

- **Communication:** All interactions secured via JWT, HTTPS, and access tokens
- **Database:** Stores users, sessions, server devices, alerts, logs

---

## 📦 Modules & Capabilities

- **Configuration Management:** Register/remove devices, execute commands, manage SSH keys, update server info
- **Network Configuration:** Interface management, routing, firewall, service restart
- **Health Monitoring:** CPU, RAM, disk, network I/O with visualizations and thresholds
- **Alert System:** Auto-detection, severity levels, filtering, admin actions
- **Log Management:** Aggregated logs, real-time streaming, filters, search
- **Resource Optimization:** Service/process management, cleanup utilities, performance suggestions
- **User Management:** Add/delete users, assign roles, profile updates

---

## 🔒 Security

- **JWT Authentication:** Secure login and API access
- **Role-Based Access:** Admins and viewers with granular permissions
- **Input & Command Validation:** Prevents injection and misuse
- **Internal Network Deployment:** Designed for secure LAN environments
- **Password Hashing & Audit Logging:** Protects credentials and tracks sensitive actions

---

## ⚙️ Configuration

Environment variables managed via `.env` files for backend, frontend, and client agents. Sensitive data (secrets, DB credentials) should be kept secure and excluded from version control.

---

## 🚀 Deployment Strategy

- Deploy backend and frontend on a central server within your secure network
- Distribute the client agent to each Linux or Windows server to be managed
- All communication is secured and access-controlled

---

## 🎯 Conclusion

SMS empowers IT teams to manage, monitor, and secure large-scale Linux server environments from a single, intuitive interface. Its modular, secure, and extensible design makes it a practical choice for institutions and enterprises seeking operational excellence.

---

## 🤝 Contributing

Contributions are welcome! Please submit issues or pull requests for new features, bug fixes, or documentation improvements.

---
