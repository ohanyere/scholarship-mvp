param(
    [string]$ComposeFile = "platform/docker-compose/docker-compose.yml",
    [string]$EnvFile = ".env",
    [string]$ServiceName = "postgres",
    [string]$MigrationsDir = "platform/db/migrations"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $EnvFile)) {
    throw "Environment file '$EnvFile' was not found. Copy .env.example to .env first."
}

Get-Content -LiteralPath $EnvFile | ForEach-Object {
    if ($_ -match '^\s*#' -or $_ -match '^\s*$') {
        return
    }

    $parts = $_ -split '=', 2
    if ($parts.Length -eq 2) {
        [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1])
    }
}

$dbUser = $env:POSTGRES_USER
$dbName = $env:POSTGRES_DB

if ([string]::IsNullOrWhiteSpace($dbUser) -or [string]::IsNullOrWhiteSpace($dbName)) {
    throw "POSTGRES_USER and POSTGRES_DB must be set in '$EnvFile'."
}

for ($attempt = 1; $attempt -le 20; $attempt++) {
    docker compose --env-file $EnvFile -f $ComposeFile exec -T $ServiceName pg_isready -U $dbUser -d $dbName *> $null
    if ($LASTEXITCODE -eq 0) {
        break
    }

    if ($attempt -eq 20) {
        throw "PostgreSQL did not become ready in time."
    }

    Start-Sleep -Seconds 2
}

docker compose --env-file $EnvFile -f $ComposeFile exec -T $ServiceName `
    psql -v ON_ERROR_STOP=1 -U $dbUser -d $dbName -c `
    "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

$migrationFiles = Get-ChildItem -LiteralPath $MigrationsDir -Filter *.sql | Sort-Object Name

foreach ($file in $migrationFiles) {
    $version = $file.Name.Replace("'", "''")
    $alreadyApplied = docker compose --env-file $EnvFile -f $ComposeFile exec -T $ServiceName `
        psql -tA -U $dbUser -d $dbName -c "SELECT 1 FROM schema_migrations WHERE version = '$version' LIMIT 1;"

    if ($alreadyApplied.Trim() -eq "1") {
        Write-Host "Skipping already-applied migration $($file.Name)"
        continue
    }

    Write-Host "Applying migration $($file.Name)"
    docker compose --env-file $EnvFile -f $ComposeFile exec -T $ServiceName `
        psql -v ON_ERROR_STOP=1 -U $dbUser -d $dbName -f "/migrations/$($file.Name)"

    docker compose --env-file $EnvFile -f $ComposeFile exec -T $ServiceName `
        psql -v ON_ERROR_STOP=1 -U $dbUser -d $dbName -c `
        "INSERT INTO schema_migrations (version) VALUES ('$version');"
}

Write-Host "Migrations complete."
