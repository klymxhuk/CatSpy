# ---------- build stage ----------
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git gcc musl-dev make bash
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN $(go env GOPATH)/bin/swag init -g sca/cmd/sca/main.go -o docs
RUN mkdir -p /app/bin \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/sca sca/cmd/sca/main.go

# ---------- runtime stage ----------
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/sca /app/sca
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
CMD ["/app/sca"]
