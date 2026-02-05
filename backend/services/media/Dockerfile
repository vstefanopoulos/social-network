FROM social-network/go-base AS build

WORKDIR /app

COPY . .

RUN go build -o media_service ./services/media/cmd/server

RUN go build -o migrate ./services/media/cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache postgresql-client

WORKDIR /app

COPY --from=build /app/media_service .
COPY --from=build /app/migrate .
COPY --from=build /app/services/media/internal/db/migrations ./migrations
# COPY --from=build /app/services/media/internal/db/seeds ./seeds

COPY services/media/entrypoint.sh /app/entrypoint.sh

# RUN chmod +x /app/seeds/seed.sh 
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./media_service"]

# CMD ["sh", "-c", "set -e; echo 'Running migrations...' && ./migrate; echo 'Running seeds...' && ./seeds/seed.sh || true; echo 'Starting service...' && exec ./media_service"]

