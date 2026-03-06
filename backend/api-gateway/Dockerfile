# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy entire project for build context
COPY . .

# Change to api-gateway directory
WORKDIR /app/backend/api-gateway

# Download dependencies (replace will work because we're in /app)
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .
# Runtime Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/backend/api-gateway/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
