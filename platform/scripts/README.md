# Scripts

This directory contains bash-based local development and operational helper scripts.

- `migrate-up.sh` waits for the local PostgreSQL container, applies ordered SQL
  migrations from `platform/db/migrations`, and records applied files in
  `schema_migrations`.
