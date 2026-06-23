# Стадия сборки: компилируем Go-приложение внутри контейнера.
FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/autoservice-crm ./cmd/server

# Финальный образ: оставляем только бинарник, шаблоны и необходимые системные пакеты.
FROM alpine:3.22
WORKDIR /app
# ttf-dejavu нужен для PDF с русским текстом.
RUN apk add --no-cache ca-certificates tzdata ttf-dejavu
COPY --from=builder /bin/autoservice-crm /app/autoservice-crm
COPY web /app/web
EXPOSE 8080
CMD ["/app/autoservice-crm"]
