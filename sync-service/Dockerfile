FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sync-service .

FROM alpine:3.19

RUN apk add --no-cache git ca-certificates

COPY --from=builder /app/sync-service /usr/local/bin/sync-service

ENTRYPOINT ["sync-service"]
