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
```

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

### 3. Install PostgreSQL

#### 🪟 On Windows:

1. Download the installer from the official site: [https://www.postgresql.org/download/windows/](https://www.postgresql.org/download/windows/)
2. Run the installer and follow the setup wizard.
3. Set the port (e.g., `8500`) and credentials (username and password) during installation.
4. After installation, ensure `pgAdmin` and the PostgreSQL service are running.

#### 🐧 On Linux (Debian/Ubuntu):

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

> Optional: Change the default port or authentication settings by editing `/etc/postgresql/<version>/main/postgresql.conf` and `pg_hba.conf`.

---

### 4. Set Up the Environment File

Create a `.env` file inside the backend directory:

```
ServerSecurityTool/backend/.env
```

Add the following content:

```env
DATABASE_URL=postgres://postgres:password@localhost:8500/SSMS?sslmode=disable
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

#### 📘 Breakdown of Important Environment Variables:

- **DATABASE_URL** → PostgreSQL connection string  
  - **postgres** → PostgreSQL username  
  - **password** → PostgreSQL password  
  - **localhost** → Database host (can be IP if remote)  
  - **8500** → PostgreSQL port  
  - **SSMS** → Database name  
  - **sslmode=disable** → Disables SSL for local/dev use

- **CLIENT_PORT** → Port used by client agents to connect  
- **CLIENT_PROTOCOL** → Communication protocol (e.g., http/https)  
- **JWT_SECRET** → Secret key for JWT token signing  
- **SERVER_PORT** → Port used by the backend server  
- **LOG_LEVEL** → Logging level (e.g., info, debug, warn, error)  

- **SMTP_HOST** → Mail server host  
- **SMTP_PORT** → Mail server port (587 for TLS)  
- **SMTP_USERNAME** → Email address for SMTP auth  
- **SMTP_PASSWORD** → SMTP password or app password  
- **SMTP_FROM** → Display name and sender email for alerts


---

### 5. Run the Backend

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

### 6. Run the Frontend

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
git clone https://github.com/KarthickRaghul/ServerManagementSuite/tree/main/client
cd client
```

Run the client:

#### 🪟 On Windows:

```powershell
go run windows\main.go
```

> Run in PowerShell **as Administrator**.

#### 🐧 On Linux:

```bash
sudo go run linux\main.go
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
