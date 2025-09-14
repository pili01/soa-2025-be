#!/bin/sh
set -e

echo "Ensuring schema is in sync..."

if [ -z "$(find prisma/migrations -maxdepth 1 -type d ! -path 'prisma/migrations' 2>/dev/null)" ]; then
  npx prisma db push
else
  npx prisma migrate deploy
fi

echo "Starting the application..."
exec "$@"