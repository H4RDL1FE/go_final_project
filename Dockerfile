# Используем официальный образ Go для сборки приложения
FROM golang:1.21.6 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальные файлы приложения
COPY . .

# Сборка приложения
RUN go build -o /go_final_project

# Используем минимальный образ Ubuntu для выполнения приложения
FROM ubuntu:latest

# Устанавливаем необходимые пакеты
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Создаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем собранное приложение из стадии сборки
COPY --from=builder /go_final_project /app/go_final_project

# Копируем файлы фронтенда
COPY ./web /app/web

# Определяем переменные окружения
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db

# Открываем порт
EXPOSE 7540

# Команда для запуска приложения
CMD ["/app/go_final_project"]
