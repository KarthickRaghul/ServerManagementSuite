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

## ⚙️ Installation Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/KarthickRaghul/ServerManagementSuite
---

### 2. Install Vite

#### 🪟 On Windows (CMD or PowerShell):

```cmd
npm install -g vite
```

#### 🐧 On Linux:

```bash
sudo npm install -g vite
```

---

### 3. Set Up the Environment File

Create a `.env` file inside the backend directory:

```
ServerSecurityTool/backend/.env
```

Add the following content:

```env
DATABASE_URL=postgres://postgres:password@localhost:8500/SSMS?sslmode=disable
```

#### 📘 Breakdown of `DATABASE_URL`:

- **postgres** → PostgreSQL **username**
- **password** → PostgreSQL **password**
- **localhost** → Database **host** (use IP address if PostgreSQL is on a different machine)
- **8500** → PostgreSQL **port**
- **SSMS** → Name of the **PostgreSQL database**
- **sslmode=disable** → Disables SSL; suitable for local development

---

### 4. Run the Backend

#### 🪟 On Windows (PowerShell as Administrator):

```powershell
cd ServerSecurityTool/backend
go run main.go
```

> ⚠️ Make sure to **Run PowerShell as Administrator** to allow access to system resources.

#### 🐧 On Linux:

```bash
cd ServerSecurityTool/backend
sudo go run main.go
```

---

### 5. Run the Frontend

In a new terminal window:

```bash
cd ServerSecurityTool/frontend
npm run dev -- --host
```

> This will launch the frontend and make it accessible over the local network.

---

## 💻 Client Setup

The client tool collects system metrics and sends data to the backend.

---

### Option 1: Clone Client on Server Machine

```bash
git clone https://github.com/your-username/ClientTool.git
cd ClientTool
```

Run the client:

#### 🪟 On Windows:

```powershell
.\client_tool.exe
```

> Run in PowerShell **as Administrator**.

#### 🐧 On Linux:

```bash
sudo ./client_tool
```

---

### Option 2: Build on Any Machine and Copy to Server

1. Build the client:

```bash
go build -o client_tool main.go
```

2. Copy the compiled binary to your server via SCP, USB, or file transfer.

3. Run on the server:

#### 🐧 Linux:

```bash
sudo ./client_tool
```

#### 🪟 Windows:

```powershell
.\client_tool.exe
```

✅ Your ServerSecurityTool stack (backend, frontend, and client) is now fully set up and running.

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
