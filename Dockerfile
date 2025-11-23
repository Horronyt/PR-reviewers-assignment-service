FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# ----------------------
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Копируем ВСЁ из builder'а — включая migrations
COPY --from=builder /app .

EXPOSE 8080
CMD ["./api"]