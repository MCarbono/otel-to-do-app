FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
EXPOSE 8888
EXPOSE 8080
ENTRYPOINT ["/app/main", "--env=production"]