# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first (for caching)
COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./

# Copy pkg go.mod for dependency resolution
COPY backend/pkg/go.mod ./pkg/

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/api-gateway/ .
COPY backend/pkg/ ./pkg/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

# Runtime Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
