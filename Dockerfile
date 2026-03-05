## Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install git for fetching Go modules hosted in git repositories
RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/game-room-manager ./cmd/game-room-manager

## Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

ENV PORT=8080
ENV APP_PORT=8080

EXPOSE 8080

COPY --from=builder /bin/game-room-manager /app/game-room-manager
COPY config ./config

USER nonroot:nonroot

ENTRYPOINT ["/app/game-room-manager"]

