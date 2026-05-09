# Инструкция по развертыванию (Deployment)

В этом руководстве описано, как упаковать приложение в Docker-контейнер и запустить его на сервере (VPS).

## Требования к серверу
- Установленный **Docker** и **Docker Compose**.
- Открытые порты 80 (HTTP) и/или 443 (HTTPS) для веб-трафика.
- Операционная система Linux (например, Ubuntu 22.04).

---

## 1. Подготовка файлов для Docker

Для упаковки приложения мы создадим `Dockerfile` в корне проекта. Он будет собирать и frontend (Next.js), и backend (Go) в один легкий контейнер.

Создайте файл `Dockerfile` в папке `C:/d/fl` (корневая папка) со следующим содержимым:

```dockerfile
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
RUN apk --no-cache add ca-certificates tzdata
# Устанавливаем временную зону по умолчанию (MSK)
ENV TZ=Europe/Moscow
WORKDIR /app

# Копируем собранный сервер
COPY --from=backend-builder /app/backend/server /app/server
# Копируем статику фронтенда
COPY --from=frontend-builder /app/frontend/out /app/frontend/out

# Указываем порт
EXPOSE 8080

# Запускаем сервер
CMD ["/app/server"]
```

## Важно: Файл .dockerignore
Для того, чтобы Docker работал корректно и не тянул лишние или заблокированные файлы (особенно если они скрыты в .gitignore), обязательно создайте файл `.dockerignore` в корне проекта со следующим содержимым:
```
node_modules/
frontend/.next/
frontend/out/
frontend/build/
.git/
.idea/
.vscode/
*.exe
backend/server
backend/server.exe
backend/credentials.json
backend/config.json
```

## 2. Настройка конфигурации на сервере

Контейнеру понадобятся ключи доступа к Google Sheets и файл конфигурации.

1. Подключитесь к вашему серверу (VPS) по SSH.
2. Исходный код проекта (папка `wedding-site`) должен лежать на сервере, например, в `/opt/wedding-app/wedding-site`.
3. В папке `wedding-site` уже есть папка `data` (`/opt/wedding-app/wedding-site/data`).
4. В эту папку `data` вам нужно поместить два файла:
   - `config.json` (с вашим Spreadsheet ID и Telegram токенами).
   - `credentials.json` (ключ от Google Service Account).

   Пример правильного `config.json` для сервера:
   ```json
   {
     "port": ":8080",
     "spreadsheetId": "ВАШ_SPREADSHEET_ID",
     "credentialsFile": "data/credentials.json",
     "tgBotToken": "ВАШ_ТОКЕН_БОТА",
     "tgChatId": "ВАШ_CHAT_ID",
     "tgCountdownChatId": "ID_ЧАТА_ДЛЯ_ОТСЧЕТА",
     "tgCountdownTopicId": "ID_ТОПИКА_ЕСЛИ_НУЖНО",
     "weddingDate": "2026-08-08T15:00:00"
   }
   ```
   **ВНИМАНИЕ:** Убедитесь, что файлы `config.json` и `credentials.json` удалены из папки `backend` во избежание конфликтов!

## 3. Запуск через Docker Compose (Рекомендуется)

Использование `docker-compose` — самый удобный способ запустить контейнер и пробросить нужные файлы.

Файл `docker-compose.yml` в папке `wedding-site` на сервере:

```yaml
version: '3.8'

services:
  wedding-app:
    build: . 
    container_name: wedding-app
    restart: unless-stopped
    ports:
      - "80:8080" # ВАЖНО: Мы пробрасываем стандартный 80 порт сервера на 8080 порт приложения!
    volumes:
      - ./data/config.json:/app/config.json
      - ./data/credentials.json:/app/data/credentials.json
```

## 4. Сборка и старт

1. Перейдите в папку с проектом (`/opt/wedding-app/wedding-site`) и выполните команду:
   ```bash
   docker compose down -v # Обязательно очищаем старые конфликтующие тома!
   docker compose up -d --build
   ```
3. Docker скачает нужные образы, соберет фронтенд, скомпилирует Go-сервер и запустит приложение в фоновом режиме.

## 5. Проверка логов (Fiber Logger)

Поскольку мы используем `Fiber` в Go, он отлично логирует все входящие запросы.
Чтобы посмотреть логи (кто заходил на сайт, какие ошибки возникли), введите команду:
```bash
docker logs -f wedding-app
```
Вы увидите подробный вывод каждого запроса и статус его выполнения.
