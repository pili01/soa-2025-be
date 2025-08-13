#!/bin/sh

set -e

echo "Running database migrations..."

npx prisma migrate deploy

echo "Migrations applied. Starting the application..."
exec "$@"