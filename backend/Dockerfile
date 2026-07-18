# PDV Backend Dockerfile
#
# Build
#   docker build -t pdv-backend .
#
# Run
#   docker run -p 8080:8080 --network host pdv-backend
#
# Running with docker-compose
#   docker compose up

FROM golang:1.26-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/api

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/.env.example ./.env

EXPOSE 8080

ENV HTTP_ADDRESS=":8080"
ENV APP_ENV="production"

CMD ["./app"]