# Database Migrations

This directory contains ordered PostgreSQL schema migrations for the Scholarship Platform MVP.

Current workflow:

- place each schema change in a new `.sql` file
- use a sortable prefix such as `0001_`, `0002_`, `0003_`
- run `make migrate-up` to apply any unapplied files

The local migration runner records applied files in a `schema_migrations` table.
