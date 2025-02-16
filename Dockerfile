FROM golang:1.23.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/shop ./cmd/shop/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/bin/shop .
COPY config/ config/
COPY migrations/ migrations/
COPY .env /app/.env
COPY ./config/application.yaml /app/config/application.yaml
EXPOSE 8080

CMD ["./shop"]