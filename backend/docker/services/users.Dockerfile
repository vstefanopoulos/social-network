FROM social-network/go-base AS build

WORKDIR /app/backend

COPY backend/ .

RUN go build -o users_service ./services/users/cmd/server

RUN go build -o migrate ./services/users/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/backend/users_service .
COPY --from=build /app/backend/migrate .
COPY --from=build /app/backend/services/users/internal/db/migrations ./migrations
COPY --from=build /app/backend/services/users/internal/db/seeds ./seeds

COPY backend/services/users/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/seeds/seed.sh && \
    chmod +x /app/entrypoint.sh

CMD ["./users_service"]
