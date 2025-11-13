# Stage 1: билд бинарника
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект внутрь контейнера
COPY . .

# Собираем бинарь из cmd/prs
RUN go build -o prs ./cmd/app

# Stage 2: минимальный образ для запуска
FROM alpine:latest

# Настройка зеркала и установка ca-certificates
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.edge.kernel.org/g' /etc/apk/repositories \
    && apk update \
    && apk add --no-cache ca-certificates

WORKDIR /app

# Копируем скомпилированный бинарь из builder
COPY --from=builder /app/prs .

# Открываем порт
EXPOSE 8080

# Команда запуска приложения
CMD ["/app/prs"]
