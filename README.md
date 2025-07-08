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

## 🚀 Quick Start (Recommended)

### Docker Deployment (Easiest Method)

**Prerequisites:**
- Docker and Docker Compose installed on your system
- Git installed

**Steps:**

1. **Clone the Repository**
   ```bash
   git clone https://github.com/KarthickRaghul/ServerManagementSuite
   cd ServerManagementSuite
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

3. **Run Setup Script**
   
   #### 🐧 On Linux:
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```
   
   #### 🪟 On Windows (PowerShell as Administrator):
   ```powershell
   .\setup.ps1
   ```

4. **Start Services**
   ```bash
   docker-compose up -d
   ```

✅ **That's it!** Your SMS is now running at `http://localhost:3000`

The setup scripts will:
- Validate Docker installation
- Configure environment variables
- Build Docker images for backend, frontend, and database
- Initialize the PostgreSQL database
- Start all services automatically

---

## 🐳 Docker Architecture

SMS uses three Docker containers:

| Container | Image | Purpose |
|-----------|-------|---------|
| **sms-backend** | `sms-backend:latest` | Go API server with JWT auth |
| **sms-frontend** | `sms-frontend:latest` | React dashboard with Vite |
| **sms-database** | `postgres:15` | PostgreSQL database |

**Docker Compose Features:**
- Automatic service dependency management
- Volume persistence for database data
- Network isolation and security
- Health checks for all services
- Auto-restart policies

---

## ⚙️ Manual Installation (Advanced)

For users who prefer manual setup or need customization beyond Docker:

### 1. Clone the Repository

```bash
git clone https://github.com/KarthickRaghul/ServerManagementSuite
```

---

### 2. Install Dependencies

#### Install Vite

**🪟 On Windows (CMD or PowerShell):**
```cmd
npm install -g vite
```

**🐧 On Linux:**
```bash
sudo npm install -g vite
```

#### Install Go
- Download from: https://golang.org/dl/
- Follow installation instructions for your OS

---

### 3. Install PostgreSQL

#### 🪟 On Windows:

1. Download the installer from: [https://www.postgresql.org/download/windows/](https://www.postgresql.org/download/windows/)
2. Run the installer and follow the setup wizard
3. Set the port (e.g., `8500`) and credentials during installation
4. Ensure `pgAdmin` and PostgreSQL service are running

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
CLIENT_PORT=8080
CLIENT_PROTOCOL=http
JWT_SECRET=your-super-secret-jwt-key-here
SERVER_PORT=8000
LOG_LEVEL=info
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=SMS Alert System <your-email@gmail.com>
```

#### 📘 Environment Variables Breakdown:

- **DATABASE_URL** → PostgreSQL connection string
  - **postgres** → PostgreSQL username
  - **password** → PostgreSQL password
  - **localhost** → Database host (can be IP if remote)
  - **8500** → PostgreSQL port
  - **SSMS** → Database name
  - **sslmode=disable** → Disables SSL for local/dev use

- **CLIENT_PORT** → Port used by client agents to connect
- **CLIENT_PROTOCOL** → Communication protocol (http/https)
- **JWT_SECRET** → Secret key for JWT token signing
- **SERVER_PORT** → Port used by the backend server
- **LOG_LEVEL** → Logging level (info, debug, warn, error)

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
go mod download
go run main.go
```

> ⚠️ Make sure to **Run PowerShell as Administrator** to allow access to system resources.

#### 🐧 On Linux:

```bash
cd ServerSecurityTool/backend
go mod download
sudo go run main.go
```

---

### 6. Run the Frontend

In a new terminal window:

```bash
cd ServerSecurityTool/frontend
npm install
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

---

## 🔧 Docker Commands Reference

### Basic Operations

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Restart a specific service
docker-compose restart sms-backend

# Rebuild images
docker-compose build

# View running containers
docker-compose ps
```

### Development Commands

```bash
# Run with live reload (development)
docker-compose -f docker-compose.dev.yml up

# Access database directly
docker exec -it sms-database psql -U postgres -d SSMS

# View backend logs
docker logs sms-backend -f

# Shell into backend container
docker exec -it sms-backend /bin/sh
```

---

## 🔑 Key Features

- **Centralized Dashboard:** Manage all servers from a unified portal—no more individual SSH sessions
- **Real-time Monitoring:** Live metrics for CPU, memory, disk, and network I/O, updated every 30 seconds with historical graphs
- **Configuration Management:** Remotely edit hostnames, network interfaces, firewall rules, and more
- **Alert System:** Automated notifications for resource thresholds (CPU, RAM, disk, network), with severity levels
- **Role-Based Access Control:** Admin (full access) and Viewer (read-only) roles, with UI tailored to each
- **Secure Communication:** JWT authentication, access tokens, input validation, and encrypted protocols (HTTPS)
- **Modular Architecture:** Easily extensible for new features and scalable to large server fleets
- **Containerized Deployment:** Docker support for easy deployment and scaling

---

## 📐 System Architecture

| Component | Technology      | Role                                         |
|-----------|----------------|----------------------------------------------|
| Frontend  | React + Vite   | User dashboard, real-time monitoring         |
| Backend   | Go + PostgreSQL| API, logic, authentication, data storage     |
| Client    | Go             | Runs on each server, collects metrics, executes commands |
| Database  | PostgreSQL     | Persistent storage for users, devices, alerts, logs |

**Communication:** All interactions secured via JWT, HTTPS, and access tokens
**Deployment:** Containerized with Docker for consistency and scalability

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
- **Container Security:** Isolated services with minimal attack surface

---

## ⚙️ Configuration

Environment variables are managed via `.env` files for backend, frontend, and client agents. Sensitive data (secrets, DB credentials) should be kept secure and excluded from version control.

**Docker Environment:**
- Environment variables are automatically loaded from `.env` file
- Database credentials are managed through Docker secrets
- SSL certificates can be mounted as volumes

---

## 🚀 Deployment Strategy

### Production Deployment

1. **Docker Deployment (Recommended)**
   - Deploy on a central server within your secure network
   - Use `docker-compose.prod.yml` for production settings
   - Configure reverse proxy (nginx) for SSL termination
   - Set up backup strategies for database volumes

2. **Manual Deployment**
   - Install dependencies on production server
   - Configure systemd services for auto-start
   - Set up log rotation and monitoring
   - Configure firewall rules

3. **Client Distribution**
   - Distribute the client agent to each Linux or Windows server
   - Use configuration management tools (Ansible, Puppet) for scale
   - Set up monitoring for client connectivity

---

## 🎯 Conclusion

SMS empowers IT teams to manage, monitor, and secure large-scale Linux server environments from a single, intuitive interface. With Docker support, deployment is now simpler than ever, while the manual installation option provides flexibility for advanced users. Its modular, secure, and extensible design makes it a practical choice for institutions and enterprises seeking operational excellence.

---

## 🤝 Contributing

Contributions are welcome! Please submit issues or pull requests for new features, bug fixes, or documentation improvements.

**Development Setup:**
```bash
# Clone and setup for development
git clone https://github.com/KarthickRaghul/ServerManagementSuite
cd ServerManagementSuite
docker-compose -f docker-compose.dev.yml up
```

---

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Docker Guide](https://hub.docker.com/_/postgres)
- [React + Vite Documentation](https://vitejs.dev/)
- [Go Documentation](https://golang.org/doc/)

---
