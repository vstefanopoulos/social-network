FROM social-network/go-base AS build

WORKDIR /app

COPY . .

RUN go build -o chat_service ./services/chat/cmd/server

RUN go build -o migrate ./services/chat/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/chat_service .
COPY --from=build /app/migrate .
COPY --from=build /app/services/chat/internal/db/migrations ./migrations
# COPY --from=build /app/services/chat/internal/db/seeds ./seeds
COPY ./services/chat/entrypoint.sh /app/entrypoint.sh


# Use a shell explicitly via JSON form
CMD ["sh", "-c", "set -e; echo 'Running migrations...' && ./migrate; echo 'Running seeds...' && ./seeds/seed.sh || true; echo 'Starting service...' && exec ./chat_service"]
