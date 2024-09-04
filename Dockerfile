FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/todos
RUN go build -o main .

FROM alpine:3.16
WORKDIR /root/
COPY --from=builder /app/todos/main .
COPY .env .
EXPOSE 8888
ENTRYPOINT ["./main", "--env=production"]