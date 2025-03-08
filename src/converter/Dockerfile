FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o converter ./cmd/api

RUN chmod +x /app/converter

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/converter ./
COPY migrations ./migrations

EXPOSE 80

CMD ["./converter", "--db-ssl=true"]
