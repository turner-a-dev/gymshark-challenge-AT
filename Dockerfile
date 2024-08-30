# Build stage
FROM golang:1.23.0 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gymshark-api

# Final stage with a minimal base image
FROM alpine:latest

COPY --from=builder /app/gymshark-api /gymshark-api
COPY --from=builder /app/static /static

EXPOSE 8080

CMD ["/gymshark-api"]