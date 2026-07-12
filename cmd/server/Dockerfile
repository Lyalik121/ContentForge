FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ffmpeg ca-certificates

COPY --from=builder /app/server .
RUN mkdir -p uploads

EXPOSE 3000
CMD ["./server"]