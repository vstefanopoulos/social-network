#!/bin/sh
set -e

echo "Running migrations..."
/app/migrate

# echo "Running seeds..."
# /app/seeds/seed.sh || true
#test, test2

echo "Starting service..."
exec ./media_service