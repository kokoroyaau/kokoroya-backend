#!/bin/sh
set -e

DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=${POSTGRES_SSL_MODE}&connect_timeout=10"

echo "Running database migrations..."
migrate -path /app/db/migrations -database "$DB_URL" up

echo "Starting api..."
exec api
