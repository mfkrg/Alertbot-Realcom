FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alertbot .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/alertbot .

COPY cron /app/cron

COPY entrypoint.sh /app/entrypoint.sh

RUN crontab /app/cron

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
