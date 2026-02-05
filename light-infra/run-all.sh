#!/usr/bin/env bash
set -euo pipefail

######################################
# Cleanup handler
######################################
cleanup() {
  echo
  echo "🛑 Stopping all services..."
  # Kill all child processes of this script
  pkill -P $$
  sleep 15
  pkill -P $$
}

trap cleanup INT TERM EXIT

######################################
# Shared environment (all services)
######################################

export SENTINEL_ADDRS=localhost:26379
export REDIS_MASTER_SET=master
export TELEMETRY_COLLECTOR_ADDR=localhost:4317


export USERS_GRPC_ADDR=localhost:50051
export POSTS_GRPC_ADDR=localhost:50052
export CHAT_GRPC_ADDR=localhost:50053
export NOTIFICATIONS_GRPC_ADDR=localhost:50054
export MEDIA_GRPC_ADDR=localhost:50055

export SHUTDOWN_TIMEOUT_SECONDS=5
export ENABLE_DEBUG_LOGS=true
export ENABLE_SIMPLE_PRINT=true
export KAFKA_BROKERS=localhost:29092

export NATS_CLUSTER=nats://localhost:4222

export GATEWAY=http://localhost:8081
export LIVE=ws://localhost:8082


######################################
# Secrets
######################################

export ENC_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=
export JWT_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=
export PASSWORD_SECRET=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=

######################################
# Root paths
######################################

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"

echo "🚀 Starting all services..."
echo "📁 Backend dir: $BACKEND_DIR"
echo

######################################
# Gateway
######################################

(
  cd "$BACKEND_DIR"
  go run ./services/gateway/cmd/main.go
) &

######################################
# Media
######################################

(
  export DATABASE_URL=postgres://postgres:secret@localhost:5437/social_media?sslmode=disable
  export MIGRATE_PATH=services/media/internal/db/migrations
  export GRPC_SERVER_PORT=:50055
  export MINIO_ENDPOINT=localhost:9000
  export MINIO_ACCESS_KEY=minioadmin
  export MINIO_SECRET_KEY=minioadmin

  cd "$BACKEND_DIR"
  go run ./services/media/cmd/server/main.go
) &

######################################
# Chat
######################################

(
  export DATABASE_URL=postgres://postgres:secret@localhost:5435/social_chat?sslmode=disable
  export MIGRATE_PATH=services/chat/internal/db/migrations
  export GRPC_SERVER_PORT=:50053

  cd "$BACKEND_DIR"
  go run ./services/chat/cmd/server/main.go 2>&1 | sed 's/^/[CHAT] /'
) &

######################################
# Posts
######################################

(
  export DATABASE_URL=postgres://postgres:secret@localhost:5434/social_posts?sslmode=disable
  export GRPC_SERVER_PORT=:50052

  cd "$BACKEND_DIR"
  go run ./services/posts/cmd/server/main.go 2>&1 | sed 's/^/[POSTS] /'
) &

######################################
# Users
######################################

(
  export DATABASE_URL=postgres://postgres:secret@localhost:5433/social_users?sslmode=disable
  export MIGRATE_PATH=services/users/internal/db/migrations
  export GRPC_SERVER_PORT=:50051

  cd "$BACKEND_DIR"
  go run ./services/users/cmd/server/main.go 2>&1 | sed 's/^/[USERS] /'
) &

######################################
# Notifications
######################################

(
  export DATABASE_URL=postgres://postgres:secret@localhost:5436/social_notifications?sslmode=disable
  export MIGRATE_PATH=services/notifications/internal/db/migrations
  export GRPC_SERVER_PORT=:50054

  cd "$BACKEND_DIR"
  go run ./services/notifications/cmd/server/main.go 2>&1 | sed 's/^/[NOTIFICATIONS] /'
) &

######################################
# Live
######################################

(

  cd "$BACKEND_DIR"
  go run ./services/live/cmd/main.go 2>&1 | sed 's/^/[LIVE] /'
) &

######################################
# Wait forever
######################################

wait
