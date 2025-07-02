
Param (
    [string]$Mode = "help"
)

function Show-Help {
    Write-Host "`n📖 Usage:"
    Write-Host "  ./setup.ps1 start   or  ./setup.ps1 -s    → Start setup"
    Write-Host "  ./setup.ps1 clean   or  ./setup.ps1 -c    → Clean Docker"
    Write-Host "  ./setup.ps1 help    or  ./setup.ps1 -h    → Show this help message`n"
    exit
}

if ($Mode -in @("help", "-h", "--help", "", $null)) {
    Show-Help
}

if ($Mode -in @("clean", "-c", "--clean")) {
    Write-Host "🧹 Cleaning up Docker containers and volumes..."
    docker compose down -v --remove-orphans

    Write-Host "🗑️  Removing old images (optional)..."
    docker rmi servermanagementsuite-backend servermanagementsuite-frontend -f | Out-Null

    Write-Host "✅ Cleanup complete. Re-run ./setup.ps1 to start fresh."
    exit
}

if (-not ($Mode -in @("start", "-s", "--start"))) {
    Write-Host "❌ Unknown argument: $Mode`n"
    Show-Help
}

# ──────────────────────────────────────────────────────────────
# 1. Detect Host IP (default route fallback)
# ──────────────────────────────────────────────────────────────
$defaultIP = (Get-NetRoute -DestinationPrefix 0.0.0.0/0 | Sort-Object RouteMetric | Select-Object -First 1 | Get-NetIPAddress).IPAddress
if (-not $defaultIP) {
    $defaultIP = "127.0.0.1"
}

# ──────────────────────────────────────────────────────────────
# 2. Prompt for IP and ports
# ──────────────────────────────────────────────────────────────
$hostIP = Read-Host "🌐 Enter host IP (default: $defaultIP)"
if (-not $hostIP) { $hostIP = $defaultIP }

$dbPort = Read-Host "🛢️  Enter DATABASE port (default: 9001)"
if (-not $dbPort) { $dbPort = 9001 }

$backendPort = Read-Host "🖥️  Enter BACKEND port (default: 9000)"
if (-not $backendPort) { $backendPort = 9000 }

# ──────────────────────────────────────────────────────────────
# 3. Create .env files
# ──────────────────────────────────────────────────────────────
$backendEnvPath = "./backend/.env"
Write-Host "⚙️  Writing backend .env to $backendEnvPath"
@"
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
"@ | Out-File -Encoding UTF8 $backendEnvPath -Force

$frontendEnvPath = "./frontend/.env"
Write-Host "⚙️  Writing frontend .env to $frontendEnvPath"
"VITE_BACKEND_URL=http://$hostIP:$backendPort" | Out-File -Encoding UTF8 $frontendEnvPath -Force

# ──────────────────────────────────────────────────────────────
# 4. Start Docker containers
# ──────────────────────────────────────────────────────────────
Write-Host "🚀 Starting Docker containers..."
$env:DB_PORT = "$dbPort"
$env:BACKEND_PORT = "$backendPort"
docker compose up -d --build

# ──────────────────────────────────────────────────────────────
# 5. Wait for DB & Init
# ──────────────────────────────────────────────────────────────
Write-Host "⏳ Waiting for database to be ready..."
Start-Sleep -Seconds 8

Write-Host "🛠️  Initializing backend database..."
if (docker exec sms-backend ./dbinit) {
    Write-Host "`n🎉 System setup complete!"
    Write-Host "🔗 Frontend:    http://$hostIP"
    Write-Host "🖥️  Backend API: http://$hostIP:$backendPort"
    Write-Host "🛢️  PostgreSQL:  Port $dbPort"
} else {
    Write-Host "❌ DB initialization failed."
    exit 1
}

