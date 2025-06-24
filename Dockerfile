FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alertbot .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/alertbot .

RUN chmod +x alertbot

CMD ["./alertbot"]
