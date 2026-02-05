FROM social-network/go-base AS build

WORKDIR /app/backend

COPY backend/ .

RUN go build -o posts_service ./services/posts/cmd/server

RUN go build -o migrate ./services/posts/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/backend/posts_service .
COPY --from=build /app/backend/migrate .
COPY --from=build /app/backend/services/posts/internal/db/migrations ./migrations
COPY --from=build /app/backend/services/posts/internal/db/seeds ./seeds

COPY backend/services/posts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/seeds/seed.sh
RUN chmod +x /app/entrypoint.sh

CMD ["./posts_service"]