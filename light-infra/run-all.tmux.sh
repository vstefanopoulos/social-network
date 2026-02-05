#!/usr/bin/env bash
set -euo pipefail

SESSION="social-backend"

######################################
# Kill existing session if exists
######################################
tmux has-session -t "$SESSION" 2>/dev/null && tmux kill-session -t "$SESSION"

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
# Paths
######################################
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"

######################################
# Create tmux session
######################################
tmux new-session -d -s "$SESSION" -n gateway

######################################
# Gateway
######################################
tmux send-keys -t "$SESSION:gateway" \
  "cd \"$BACKEND_DIR\" && go run ./services/gateway/cmd/main.go" C-m

######################################
# Media
######################################
tmux new-window -t "$SESSION" -n media
tmux send-keys -t "$SESSION:media" "
  export DATABASE_URL=postgres://postgres:secret@localhost:5437/social_media?sslmode=disable
  export MIGRATE_PATH=services/media/internal/db/migrations
  export GRPC_SERVER_PORT=:50055
  export MINIO_ENDPOINT=localhost:9000
  export MINIO_ACCESS_KEY=minioadmin
  export MINIO_SECRET_KEY=minioadmin
  cd \"$BACKEND_DIR\" && go run ./services/media/cmd/server/main.go
" C-m

######################################
# Chat
######################################
tmux new-window -t "$SESSION" -n chat
tmux send-keys -t "$SESSION:chat" "
  export DATABASE_URL=postgres://postgres:secret@localhost:5435/social_chat?sslmode=disable
  export MIGRATE_PATH=services/chat/internal/db/migrations
  export GRPC_SERVER_PORT=:50053
  cd \"$BACKEND_DIR\" && go run ./services/chat/cmd/server/main.go
" C-m

######################################
# Posts
######################################
tmux new-window -t "$SESSION" -n posts
tmux send-keys -t "$SESSION:posts" "
  export DATABASE_URL=postgres://postgres:secret@localhost:5434/social_posts?sslmode=disable
  export GRPC_SERVER_PORT=:50052
  cd \"$BACKEND_DIR\" && go run ./services/posts/cmd/server/main.go
" C-m

######################################
# Users
######################################
tmux new-window -t "$SESSION" -n users
tmux send-keys -t "$SESSION:users" "
  export DATABASE_URL=postgres://postgres:secret@localhost:5433/social_users?sslmode=disable
  export MIGRATE_PATH=services/users/internal/db/migrations
  export GRPC_SERVER_PORT=:50051
  cd \"$BACKEND_DIR\" && go run ./services/users/cmd/server/main.go
" C-m

######################################
# Notifications
######################################
tmux new-window -t "$SESSION" -n notifications
tmux send-keys -t "$SESSION:notifications" "
  export DATABASE_URL=postgres://postgres:secret@localhost:5436/social_notifications?sslmode=disable
  export MIGRATE_PATH=services/notifications/internal/db/migrations
  export GRPC_SERVER_PORT=:50054
  cd \"$BACKEND_DIR\" && go run ./services/notifications/cmd/server/main.go
" C-m

######################################
# Live
######################################
tmux new-window -t "$SESSION" -n live
tmux send-keys -t "$SESSION:live" "
  cd \"$BACKEND_DIR\" && go run ./services/live/cmd/main.go
" C-m

######################################
# Attach
######################################
tmux attach -t "$SESSION"
