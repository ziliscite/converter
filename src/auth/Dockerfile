FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o auth ./cmd/api

RUN chmod +x /app/auth

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/auth ./
COPY migrations ./migrations

EXPOSE 80

CMD ["./auth", "--db-ssl=true"]
