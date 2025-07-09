#!/bin/bash

set -e

print_help() {
  echo ""
  echo "📦 Server Management Suite Setup Script"
  echo ""
  echo "Usage:"
  echo "  ./setup.sh start      or  -s     Run the full setup (build, env, docker)"
  echo "  ./setup.sh clean      or  -c     Clean up all containers, volumes, and images"
  echo "  ./setup.sh help       or  -h     Show this help message"
  echo "  ./setup.sh exit       or  -e     This will stop the running containers"
  echo ""
  exit 0
}

# ──────────────────────────────────────────────
# 0. Argument Parsing
# ──────────────────────────────────────────────
case "$1" in
start | -s)
  echo "📦 Starting Server Management Suite Setup"
  ;;
clean | -c)
  echo "🧹 Cleaning up Docker containers and volumes..."
  docker compose down -v --remove-orphans

  echo "🗑️  Removing old images (optional)..."
  docker rmi servermanagementsuite-backend servermanagementsuite-frontend -f || true

  echo "✅ Cleanup complete. Re-run './setup.sh start' to start fresh."
  exit 0
  ;;
exit | -e)
  echo "🛑 Stopping running containers..."
  docker compose stop
  echo "✅ Containers stopped. You can resume later with 'docker compose start'."
  exit 0
  ;;
help | -h | "" | *)
  print_help
  ;;
esac

# ──────────────────────────────────────────────
# 1. Detect Host IPs and Let User Choose
# ──────────────────────────────────────────────

# Get all non-loopback IPv4 addresses
mapfile -t IP_LIST < <(ip -4 addr show | awk '/inet / {print $2}' | cut -d/ -f1 | grep -v '^127\.')

# Fallback to localhost if nothing found
if [[ ${#IP_LIST[@]} -eq 0 ]]; then
  IP_LIST=("127.0.0.1")
fi

# Detect the default/active interface IP
DEFAULT_IP=$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if ($i=="src") print $(i+1)}')
if [[ -z "$DEFAULT_IP" ]]; then
  DEFAULT_IP="${IP_LIST[0]}"
fi

# Find index of DEFAULT_IP in IP_LIST
DEFAULT_IDX=1
for idx in "${!IP_LIST[@]}"; do
  if [[ "${IP_LIST[$idx]}" == "$DEFAULT_IP" ]]; then
    DEFAULT_IDX=$((idx + 1))
    break
  fi
done

echo ""
echo "🌐 Available host IP addresses:"
for i in "${!IP_LIST[@]}"; do
  n=$((i + 1))
  if [[ "$n" -eq "$DEFAULT_IDX" ]]; then
    echo "  $n) ${IP_LIST[$i]}   (default/active)"
  else
    echo "  $n) ${IP_LIST[$i]}"
  fi
done

read -p "Select host IP [1-${#IP_LIST[@]}] (default: $DEFAULT_IDX): " IP_CHOICE
IP_CHOICE=${IP_CHOICE:-$DEFAULT_IDX}

# Validate input
while ! [[ "$IP_CHOICE" =~ ^[0-9]+$ ]] || ((IP_CHOICE < 1 || IP_CHOICE > ${#IP_LIST[@]})); do
  echo "❌ Invalid choice."
  read -p "Select host IP [1-${#IP_LIST[@]}] (default: $DEFAULT_IDX): " IP_CHOICE
  IP_CHOICE=${IP_CHOICE:-$DEFAULT_IDX}
done

HOST_IP="${IP_LIST[$((IP_CHOICE - 1))]}"
echo "✅ Using IP: $HOST_IP"
echo ""

DEFAULT_DB_PORT=9001
DEFAULT_BACKEND_PORT=9000

read -p "🛢️  Enter DATABASE port (default: $DEFAULT_DB_PORT): " DB_PORT
DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}
while ! [[ "$DB_PORT" =~ ^[0-9]+$ ]]; do
  echo "❌ Invalid DB port."
  read -p "🛢️  Enter DATABASE port (default: $DEFAULT_DB_PORT): " DB_PORT
  DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}
done

read -p "🖥️  Enter BACKEND port (default: $DEFAULT_BACKEND_PORT): " BACKEND_PORT
BACKEND_PORT=${BACKEND_PORT:-$DEFAULT_BACKEND_PORT}
while ! [[ "$BACKEND_PORT" =~ ^[0-9]+$ ]]; do
  echo "❌ Invalid BACKEND port."
  read -p "🖥️  Enter BACKEND port (default: $DEFAULT_BACKEND_PORT): " BACKEND_PORT
  BACKEND_PORT=${BACKEND_PORT:-$DEFAULT_BACKEND_PORT}
done

# ──────────────────────────────────────────────
# 2. Generate backend .env
# ──────────────────────────────────────────────
BACKEND_ENV_FILE="./backend/.env"
echo "⚙️  Generating backend .env at $BACKEND_ENV_FILE"
mkdir -p ./backend

cat >"$BACKEND_ENV_FILE" <<EOF
DATABASE_URL=postgres://admin:admin@sms-db:5432/smsdb?sslmode=disable
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
EOF
echo "✅ Backend .env created"

# ──────────────────────────────────────────────
# 3. Generate frontend .env
# ──────────────────────────────────────────────
FRONTEND_ENV_FILE="./frontend/.env"
echo "⚙️  Generating frontend .env at $FRONTEND_ENV_FILE"
mkdir -p ./frontend

cat >"$FRONTEND_ENV_FILE" <<EOF
VITE_BACKEND_URL=http://$HOST_IP:$BACKEND_PORT
EOF
echo "✅ Frontend .env created"

# ──────────────────────────────────────────────
# 4. Export vars for Docker Compose
# ──────────────────────────────────────────────
export DB_PORT=$DB_PORT
export BACKEND_PORT=$BACKEND_PORT

# ──────────────────────────────────────────────
# 5. Start Docker
# ──────────────────────────────────────────────
echo "🚀 Starting Docker containers..."
docker compose up -d --build

# ──────────────────────────────────────────────
# 6. Wait and Init DB
# ──────────────────────────────────────────────
echo "⏳ Waiting for database to be ready..."
sleep 8

echo "🛠️  Initializing backend database..."
if docker exec sms-backend ./dbinit; then
  echo ""
  echo "🎉 System setup complete!"
  echo "🔗 Frontend:    http://$HOST_IP"
  echo "🖥️  Backend API: http://$HOST_IP:$BACKEND_PORT"
  echo "🛢️  PostgreSQL:  Port $DB_PORT"
else
  echo "❌ DB initialization failed."
  exit 1
fi
