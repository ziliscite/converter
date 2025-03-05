FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o gateway ./cmd/api

RUN chmod +x /app/gateway

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/gateway ./

EXPOSE 80

CMD ["./gateway"]
