#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"

compose_file="$repo_root/platform/docker-compose/docker-compose.yml"
env_file="$repo_root/.env"
service_name="postgres"
migrations_dir="$repo_root/platform/db/migrations"

usage() {
  cat <<USAGE
Usage: $0 [options]

Options:
  --compose-file PATH   Docker Compose file to use
  --env-file PATH       Environment file to load
  --service-name NAME   PostgreSQL Compose service name
  --migrations-dir DIR  Directory containing ordered .sql migrations
  -h, --help            Show this help text
USAGE
}

absolute_path() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s\n' "$repo_root/$1" ;;
  esac
}

sql_quote() {
  printf "%s" "$1" | sed "s/'/''/g"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --compose-file)
      compose_file="$(absolute_path "$2")"
      shift 2
      ;;
    --env-file)
      env_file="$(absolute_path "$2")"
      shift 2
      ;;
    --service-name)
      service_name="$2"
      shift 2
      ;;
    --migrations-dir)
      migrations_dir="$(absolute_path "$2")"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "$env_file" ]]; then
  echo "Environment file '$env_file' was not found. Copy .env.example to .env first." >&2
  exit 1
fi

if [[ ! -d "$migrations_dir" ]]; then
  echo "Migrations directory '$migrations_dir' was not found." >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker was not found on PATH." >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$env_file"
set +a

: "${POSTGRES_USER:?POSTGRES_USER must be set in $env_file}"
: "${POSTGRES_DB:?POSTGRES_DB must be set in $env_file}"

compose=(docker compose --env-file "$env_file" -f "$compose_file")

echo "Waiting for PostgreSQL to become ready..."
for attempt in {1..20}; do
  if "${compose[@]}" exec -T "$service_name" pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    break
  fi

  if [[ "$attempt" -eq 20 ]]; then
    echo "PostgreSQL did not become ready in time." >&2
    exit 1
  fi

  sleep 2
done

"${compose[@]}" exec -T "$service_name" \
  psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
  "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

shopt -s nullglob
migration_files=("$migrations_dir"/*.sql)

if [[ "${#migration_files[@]}" -eq 0 ]]; then
  echo "No migration files found in $migrations_dir."
  exit 0
fi

mapfile -t migration_files < <(printf '%s\n' "${migration_files[@]}" | sort)

for migration_path in "${migration_files[@]}"; do
  migration_name="$(basename "$migration_path")"
  migration_version="$(sql_quote "$migration_name")"

  already_applied="$("${compose[@]}" exec -T "$service_name" \
    psql -tA -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
    "SELECT 1 FROM schema_migrations WHERE version = '$migration_version' LIMIT 1;")"

  if [[ "$already_applied" == "1" ]]; then
    echo "Skipping already-applied migration $migration_name"
    continue
  fi

  echo "Applying migration $migration_name"
  "${compose[@]}" exec -T "$service_name" \
    psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/migrations/$migration_name"

  "${compose[@]}" exec -T "$service_name" \
    psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
    "INSERT INTO schema_migrations (version) VALUES ('$migration_version');"
done

echo "Migrations complete."
