#!/bin/sh
set -e

echo "Seeding dev database..."

# Make sure DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
  echo "Error: DATABASE_URL is not set"
  exit 1
fi

psql "$DATABASE_URL" -f /app/seeds/seed.sql

echo "Seeding complete!"
