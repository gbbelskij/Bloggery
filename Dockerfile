FROM golang:1.24.1 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server cmd/main.go

FROM debian:stable-slim

# Создаём непривилегированного пользователя
RUN useradd -m -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/config /app/config
COPY entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh
# Даём права на файлы
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
