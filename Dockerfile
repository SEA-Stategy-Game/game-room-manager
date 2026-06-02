## Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install git for fetching Go modules hosted in git repositories
RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/game-room-manager ./cmd/game-room-manager

# Create a directory for the database
RUN mkdir /data
FROM gcr.io/distroless/base-debian12

WORKDIR /app

ENV PORT=8080
ENV APP_PORT=8080

EXPOSE 8080

COPY --from=builder /bin/game-room-manager /app/game-room-manager
COPY config ./config
# Copy the data directory and set ownership to the nonroot user
COPY --from=builder --chown=nonroot:nonroot /data /data

USER nonroot:nonroot

ENTRYPOINT ["/app/game-room-manager"]
