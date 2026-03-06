# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first
COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./

# Copy pkg directory
COPY backend/pkg ./pkg

# Download dependencies (vendor approach)
RUN go mod download && go mod vendor

# Copy source code
COPY backend/api-gateway/ .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .
# Runtime Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
