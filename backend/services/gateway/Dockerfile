# ---- Build stage ----
FROM social-network/go-base AS build

WORKDIR /app

COPY . .

RUN go build -o api_gateway ./services/gateway/cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/api_gateway .

CMD ["./api_gateway"]