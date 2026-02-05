FROM social-network/go-base AS build

WORKDIR /app/backend

COPY backend/ .

RUN go build -o notifications_service ./services/notifications/cmd/server

RUN go build -o migrate ./services/notifications/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/backend/notifications_service .
COPY --from=build /app/backend/migrate .
COPY --from=build /app/backend/services/notifications/internal/db/migrations ./migrations

COPY backend/services/notifications/entrypoint.sh /app/entrypoint.sh

CMD ["./notifications_service"]
