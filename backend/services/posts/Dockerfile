FROM social-network/go-base AS build

WORKDIR /app

COPY . .

RUN go build -o posts_service ./services/posts/cmd/server

RUN go build -o migrate ./services/posts/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/posts_service .
COPY --from=build /app/migrate .
COPY --from=build /app/services/posts/internal/db/migrations ./migrations
COPY --from=build /app/services/posts/internal/db/seeds ./seeds

COPY services/posts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/seeds/seed.sh
RUN chmod +x /app/entrypoint.sh

# ENTRYPOINT ["/app/entrypoint.sh"]
# CMD ["./posts_service"]
CMD ["sh", "-c", "set -e; echo 'Running migrations...' && ./migrate; echo 'Running seeds...' && ./seeds/seed.sh || true; echo 'Starting service...' && exec ./posts_service"]
