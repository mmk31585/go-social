set -euo pipefail

source .env

SUPERUSER="postgres://admin:adminpassword@localhost:5435/postgres?sslmode=disable"

echo "Creating database and user if they don't exist..."
psql "$SUPERUSER" <<-SQL
    SELECT 'CREATE DATABASE social'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'social')\gexec
    SELECT 'CREATE ROLE admin LOGIN PASSWORD ''adminpassword'''
    WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'admin')\gexec
    GRANT ALL PRIVILEGES ON DATABASE social TO admin;
SQL

echo "Running migrations..."
migrate -path=./cmd/migrate/migrations -database="$DB_ADDR" up