# Используем официальный образ Go
FROM golang:1.24-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./
RUN go mod tidy

# Копируем весь проект в контейнер
COPY . .

# Собираем приложение
RUN go build -o /calculator_with_auth cmd/server/main.go

# Указываем команду для запуска контейнера
CMD ["/calculator_with_auth"]