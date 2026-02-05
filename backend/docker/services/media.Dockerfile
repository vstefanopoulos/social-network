FROM social-network/go-base AS build

WORKDIR /app/backend

COPY backend/ .

RUN go build -o media_service ./services/media/cmd/server

RUN go build -o migrate ./services/media/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/backend/media_service .
COPY --from=build /app/backend/migrate .
COPY --from=build /app/backend/services/media/internal/db/migrations ./migrations
# COPY --from=build /app/services/media/internal/db/seeds ./seeds

COPY backend/services/media/entrypoint.sh /app/entrypoint.sh

CMD ["./media_service"]