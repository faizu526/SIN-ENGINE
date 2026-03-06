# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy entire project first (needed for go mod with local replace)
COPY . /app

# Download dependencies
RUN cd /app/backend/api-gateway && go mod download

# Build the binary
RUN cd /app/backend/api-gateway && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .
# Runtime Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/backend/api-gateway/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
