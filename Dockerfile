# Сборка Frontend (Next.js)
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Сборка Backend (Go)
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# Сборка исполняемого файла
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Финальный легковесный образ
FROM alpine:latest
# Устанавливаем корневые сертификаты (обязательно для работы Telegram API / HTTPS)
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Копируем собранный сервер
COPY --from=backend-builder /app/backend/server /app/server
# Копируем статику фронтенда
COPY --from=frontend-builder /app/frontend/out /app/frontend/out

# Указываем порт
EXPOSE 8080

# Запускаем сервер
CMD ["/app/server"]