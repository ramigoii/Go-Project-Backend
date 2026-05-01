FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]