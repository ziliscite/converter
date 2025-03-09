FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o mailer ./cmd/api

RUN chmod +x /app/mailer

FROM alpine:latest

WORKDIR /app

COPY --from=builder app/mailer ./

EXPOSE 80

CMD ["./mailer"]
