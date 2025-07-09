Param (
    [string]$Mode = "help"
)

function Show-Help {
    Write-Host ""
    Write-Host "[Server Management Suite Setup Script]"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  ./setup.ps1 start       → Start setup"
    Write-Host "  ./setup.ps1 clean       → Clean Docker"
    Write-Host "  ./setup.ps1 exit        → Stop containers"
    Write-Host "  ./setup.ps1 help        → Show this help"
    Write-Host ""
    exit 0
}

if ($Mode -eq "start" -or $Mode -eq "-s") {
    Write-Host "[*] Starting Server Management Suite Setup"
} elseif ($Mode -eq "clean" -or $Mode -eq "-c") {
    Write-Host "[*] Cleaning up Docker containers and volumes..."
    docker compose down -v --remove-orphans

    Write-Host "[*] Removing old images..."
    docker rmi servermanagementsuite-backend servermanagementsuite-frontend -f 2>$null
    Write-Host "[✓] Cleanup complete. Re-run './setup.ps1 start' to start fresh."
    exit 0
} elseif ($Mode -eq "exit" -or $Mode -eq "-e") {
    Write-Host "[*] Stopping running containers..."
    docker compose stop
    Write-Host "[✓] Containers stopped. Resume later with 'docker compose start'"
    exit 0
} else {
    Show-Help
}

# ───────────────────────────────
# Select or Enter Host IP
# ───────────────────────────────
$ipOptions = @(Get-NetIPAddress -AddressFamily IPv4 | Where-Object {
    $_.IPAddress -notlike '169.*' -and $_.InterfaceAlias -notmatch 'Loopback|Virtual|vEthernet|WSL|Hyper-V|VPN|Docker'
})


if ($ipOptions.Count -eq 0) {
    Write-Host "[X] No valid host IPs found automatically. Please enter manually."
    $HOST_IP = Read-Host "Enter Host IP"
} else {
    $i = 0
    Write-Host "`nAvailable Network Interfaces:"
    $ipOptions | ForEach-Object {
        Write-Host "$i) $($_.IPAddress)  ($($_.InterfaceAlias))"
        $i++
    }

    Write-Host "m) Manually enter IP"

    $choice = Read-Host "Select IP (default 0 or enter 'm' to input manually)"
    if ([string]::IsNullOrWhiteSpace($choice)) { $choice = "0" }

    if ($choice -eq 'm') {
        $HOST_IP = Read-Host "Enter Host IP manually"
    } elseif ($choice -match '^\d+$') {
        $index = [int]$choice
        if ($index -ge 0 -and $index -lt $ipOptions.Count) {
            $HOST_IP = $ipOptions[$index].IPAddress
        } else {
            Write-Host "[X] Invalid index selected. Exiting."
            exit 1
        }
    } else {
        Write-Host "[X] Invalid choice. Exiting."
        exit 1
    }
}

# ───────────────────────────────
# Read Ports
# ───────────────────────────────
$DEFAULT_DB_PORT = 9001
$DEFAULT_BACKEND_PORT = 9000

$DB_PORT = Read-Host "Enter DATABASE port (default: $DEFAULT_DB_PORT)"
if (-not ($DB_PORT -match '^\d+$')) { $DB_PORT = $DEFAULT_DB_PORT }

$BACKEND_PORT = Read-Host "Enter BACKEND port (default: $DEFAULT_BACKEND_PORT)"
if (-not ($BACKEND_PORT -match '^\d+$')) { $BACKEND_PORT = $DEFAULT_BACKEND_PORT }

# ───────────────────────────────
# Generate backend .env
# ───────────────────────────────
$BACKEND_ENV_FILE = "./backend/.env"
Write-Host "[*] Generating backend .env at $BACKEND_ENV_FILE"
New-Item -Path "./backend" -ItemType Directory -Force | Out-Null

$backendEnvContent = @"
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
SMTP_FROM='SMS Alerts <servermanagementcit@gmail.com>'
"@

$backendEnvContent | Set-Content -Path $BACKEND_ENV_FILE
Write-Host "[✓] Backend .env created"

# ───────────────────────────────
# Generate frontend .env
# ───────────────────────────────
$FRONTEND_ENV_FILE = "./frontend/.env"
Write-Host "[*] Generating frontend .env at $FRONTEND_ENV_FILE"
New-Item -Path "./frontend" -ItemType Directory -Force | Out-Null

"VITE_BACKEND_URL=http://${HOST_IP}:${BACKEND_PORT}" | Set-Content -Path $FRONTEND_ENV_FILE
Write-Host "[✓] Frontend .env created"

# ───────────────────────────────
# Export ports as env vars (optional for Compose)
# ───────────────────────────────
$env:DB_PORT = $DB_PORT
$env:BACKEND_PORT = $BACKEND_PORT

# ───────────────────────────────
# Start Docker
# ───────────────────────────────
Write-Host "[*] Starting Docker containers..."
docker compose up -d --build

# ───────────────────────────────
# Wait and Init DB
# ───────────────────────────────
Write-Host "[*] Waiting for DB to be ready..."
Start-Sleep -Seconds 8

Write-Host "[*] Initializing backend database..."
$dockerExecResult = docker exec sms-backend ./dbinit
$initSuccess = $LASTEXITCODE -eq 0

if ($initSuccess) {
    Write-Host ""
    Write-Host "[✓] System setup complete!"
    Write-Host "Frontend:    http://${HOST_IP}"
    Write-Host "Backend API: http://${HOST_IP}:${BACKEND_PORT}"
    Write-Host "PostgreSQL:  Port ${DB_PORT}"
} else {
    Write-Host "[X] DB initialization failed."
    exit 1
}
