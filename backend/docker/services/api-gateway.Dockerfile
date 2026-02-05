FROM social-network/go-base AS build

WORKDIR /app/backend

COPY backend/ .

RUN go build -o api_gateway ./services/gateway/cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/backend/api_gateway .

CMD ["./api_gateway"]