#!/bin/sh
set -e

echo "Running migrations..."
/app/migrate

echo "Running seeds..."
/app/seeds/seed.sh || true

echo "Starting service..."
exec ./users_service