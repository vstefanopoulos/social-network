# Run from light-infra
# make docker-up
# make run-migrations
# make run-all

.PHONY: \
	run-services run-all \
	run-api run-media run-chat run-posts run-users run-notifications\
	run-migrations \
	migrate-media migrate-users migrate-posts migrate-chat migrate-notifications\
	create-network \
	docker-up docker-down \
	tmux stop

######################################
# Shared environment (all services)
######################################

export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=admin
export REDIS_MASTER=mymaster
export REDIS_DB=0
export SENTINEL_ADDRS=localhost:26379

export NATS_CLUSTER=nats://ruser:T0pS3cr3t@localhost:4222,nats://ruser:T0pS3cr3t@localhost:4222

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



######################################
# Top-level commands
######################################

run-services: run-migrations run-all

######################################
# Run ALL services (parallel, blocking)
######################################


build-all:
	@echo "Starting building..."
	trap 'kill 0' INT TERM; \
	OUTDIR="$(CURDIR)"; \
	cd ../backend && \
	go build -o "$$OUTDIR/gateway.social" ./services/gateway/cmd/main.go; \
	go build -o "$$OUTDIR/media.social" ./services/media/cmd/server/main.go; \
	go build -o "$$OUTDIR/chat.social" ./services/chat/cmd/server/main.go; \
	go build -o "$$OUTDIR/posts.social" ./services/posts/cmd/server/main.go; \
	go build -o "$$OUTDIR/users.social" ./services/users/cmd/server/main.go; \
 	go build -o "$$OUTDIR/notifs.social" ./services/notifications/cmd/server/main.go; \
 	go build -o "$$OUTDIR/live.social" ./services/live/cmd/main.go; \


run-built:
	@echo "Starting all services..."
	@trap 'kill 0' INT TERM; \
	\

	ENC_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	JWT_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	PASSWORD_SECRET=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	TELEMETRY_COLLECTOR_ADDR=localhost:4317 \
	./gateway.social 2>&1 | sed 's/^/[API] /' & \
	\
	DATABASE_URL=postgres://postgres:secret@localhost:5437/social_media?sslmode=disable \
	MIGRATE_PATH=services/media/internal/db/migrations \
	GRPC_SERVER_PORT=:50055 \
	MINIO_ENDPOINT=localhost:9000 \
	MINIO_ACCESS_KEY=minioadmin \
	MINIO_SECRET_KEY=minioadmin \
	KAFKA_BROKERS=localhost:29092 \
	./media.social & \
	\
	DATABASE_URL=postgres://postgres:secret@localhost:5435/social_chat?sslmode=disable \
	MIGRATE_PATH=services/chat/internal/db/migrations \
	GRPC_SERVER_PORT=:50053 \
	NATS_HOST=localhost \
	KAFKA_BROKERS=localhost:29092 \
	./chat.social 2>&1 | sed 's/^/[CHAT] /' & \
	\
	DATABASE_URL=postgres://postgres:secret@localhost:5434/social_posts?sslmode=disable \
	GRPC_SERVER_PORT=:50052 \
	KAFKA_BROKERS=localhost:29092 \
	./posts.social 2>&1 | sed 's/^/[POSTS] /' & \
	\
	DATABASE_URL=postgres://postgres:secret@localhost:5433/social_users?sslmode=disable \
	MIGRATE_PATH=services/users/internal/db/migrations \
	GRPC_SERVER_PORT=:50051 \
	KAFKA_BROKERS=localhost:29092 \
	./users.social 2>&1 | sed 's/^/[USERS] /' & \
	\
	DATABASE_URL=postgres://postgres:secret@localhost:5436/social_notifications?sslmode=disable \
	MIGRATE_PATH=services/notifications/internal/db/migrations \
	GRPC_SERVER_PORT=:50054 \
	KAFKA_BROKERS=localhost:29092 \
	./notifs.social 2>&1 | sed 's/^/[NOTIFICATIONS] /' & \
	\
	ENC_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	JWT_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	PASSWORD_SECRET=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	NATS_HOST=localhost \
	./live.social 2>&1 | sed 's/^/[LIVE] /' & \
	\
	wait

######################################
# Individual services
######################################

run-api:
	ENC_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	JWT_KEY=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	PASSWORD_SECRET=a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28= \
	go run -C ../backend ./services/gateway/cmd/main.go

run-users:
	DATABASE_URL=postgres://postgres:secret@localhost:5433/social_users?sslmode=disable \
	MIGRATE_PATH=services/users/internal/db/migrations \
	GRPC_SERVER_PORT=:50051 \
	go run -C ../backend ./services/users/cmd/server/main.go 2>&1 | sed 's/^/[USERS] /'

run-posts:
	DATABASE_URL=postgres://postgres:secret@localhost:5434/social_posts?sslmode=disable \
	GRPC_SERVER_PORT=:50052 \
	go run -C ../backend ./services/posts/cmd/server/main.go 2>&1 | sed 's/^/[POSTS] /'

run-chat:
	DATABASE_URL=postgres://postgres:secret@localhost:5435/social_chat?sslmode=disable \
	MIGRATE_PATH=services/chat/internal/db/migrations \
	GRPC_SERVER_PORT=:50053 \
	go run -C ../backend ./services/chat/cmd/server/main.go 2>&1 | sed 's/^/[CHAT] /'

run-media:
	DATABASE_URL=postgres://postgres:secret@localhost:5437/social_media?sslmode=disable \
	MIGRATE_PATH=services/media/internal/db/migrations \
	GRPC_SERVER_PORT=:50055 \
	MINIO_ENDPOINT=localhost:9000 \
	MINIO_ACCESS_KEY=minioadmin \
	MINIO_SECRET_KEY=minioadmin \
	go run -C ../backend ./services/media/cmd/server/main.go

run-notifications:
	DATABASE_URL=postgres://postgres:secret@localhost:5436/social_notifications?sslmode=disable \
	MIGRATE_PATH=services/notifications/internal/db/migrations \
	GRPC_SERVER_PORT=:50054 \
	KAFKA_BROKERS=localhost:29092 \
	go run -C ../backend ./services/notifications/cmd/server/main.go 2>&1 | sed 's/^/[NOTIFICATIONS] /'

######################################
# Migrations
######################################

run-migrations: migrate-media migrate-users migrate-posts migrate-chat migrate-notifications

migrate-media:
	DATABASE_URL=postgres://postgres:secret@localhost:5437/social_media?sslmode=disable \
	MIGRATE_PATH=services/media/internal/db/migrations \
	go run -C ../backend ./services/media/cmd/migrate

migrate-chat:
	DATABASE_URL=postgres://postgres:secret@localhost:5435/social_chat?sslmode=disable \
	MIGRATE_PATH=services/chat/internal/db/migrations \
	go run -C ../backend ./services/chat/cmd/migrate

migrate-users:
	DATABASE_URL=postgres://postgres:secret@localhost:5433/social_users?sslmode=disable \
	MIGRATE_PATH=services/users/internal/db/migrations \
	go run -C ../backend ./services/users/cmd/migrate

migrate-posts:
	DATABASE_URL=postgres://postgres:secret@localhost:5434/social_posts?sslmode=disable \
	MIGRATE_PATH=services/posts/internal/db/migrations \
	go run -C ../backend ./services/posts/cmd/migrate

migrate-notifications:
	DATABASE_URL=postgres://postgres:secret@localhost:5436/social_notifications?sslmode=disable \
	MIGRATE_PATH=services/notifications/internal/db/migrations \
	go run -C ../backend ./services/notifications/cmd/migrate

######################################
# Docker helpers
######################################

docker-up:
	$(MAKE) create-network
	docker compose -f docker-compose.infra.yml up -d

docker-down:
	docker compose -f docker-compose.infra.yml down

create-network:
	@docker network inspect social-network >/dev/null 2>&1 || \
	docker network create social-network

boom:
	$(MAKE) docker-up
	$(MAKE) run-migrations
	./run-all.sh

######################################
# Run light
######################################

tmux:
	$(MAKE) docker-up
	$(MAKE) run-migrations
	./run-all.tmux.sh

stop:
	tmux kill-session -t social-backend
	docker compose -f docker-compose.infra.yml down

