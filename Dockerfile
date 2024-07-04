FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o ./bot ./cmd/bot


FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/bot .
EXPOSE 8080
ENTRYPOINT ["./bot"]

