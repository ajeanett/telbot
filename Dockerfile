FROM golang:1.21-alpine

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение из cmd/bot
RUN go build -o bot ./cmd/bot

EXPOSE 8080

CMD ["./bot"]