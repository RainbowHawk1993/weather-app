FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o weatherapi_service ./cmd/weatherapi_service/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY ./migrations ./migrations
COPY --from=builder /app/weatherapi_service .
COPY ./web ./web

EXPOSE 8080
CMD ["./weatherapi_service"]
