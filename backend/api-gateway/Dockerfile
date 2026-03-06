FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY backend/api-gateway/go.mod backend/api-gateway/go.sum ./
RUN go mod download

COPY backend/api-gateway/ .
RUN go build -o api-gateway .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
