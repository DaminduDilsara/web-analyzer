# ---------- Stage 1: Build ----------
FROM golang:1.24.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the app
RUN go build -o app .

# ---------- Stage 2: Run ----------
FROM alpine:latest

# Install CA certificates (for HTTPS)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary and config file
COPY --from=builder /app/app .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/analyzer.html .

EXPOSE 8080

ENTRYPOINT ["./app"]
