FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./bot ./cmd/bot


FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/bot .
EXPOSE 8080
ENTRYPOINT ["./bot"]

