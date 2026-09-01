
# --- Stage 1: build -----------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/server

# --- Stage 2: runtime -----------------------------------------------------

FROM alpine:3.18 AS production

WORKDIR /app

RUN addgroup -S app && adduser -S -D -H -G app app

COPY --from=builder --chown=app:app /app/order-service .

USER app

EXPOSE 8082

CMD ["./order-service"]
