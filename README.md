  # game-room-manager

Minimal Go HTTP service template using:

- Viper for configuration
- Zap for structured logging
- Docker for containerization

## Requirements

- Go 1.25+
- Docker (optional, for container builds)

## Configuration

Configuration is loaded with Viper in the following precedence:

1. Environment variables (`APP_PORT`, `APP_ENV`, `APP_LOG_LEVEL`)
2. YAML file `config/config.yaml`
3. Hard-coded defaults

### Config keys

- `port` / `APP_PORT` (int): HTTP port.
  - **Default**: `8080`
- `env` / `APP_ENV` (string): Environment name (`local`, `dev`, `prod`, etc.).
  - **Default**: `local`
- `log_level` / `APP_LOG_LEVEL` (string): Log level (`debug`, `info`, `warn`, `error`).
  - **Default**: `info`

Example `config/config.yaml`:

```yaml
port: 8080
env: local
log_level: info
```

> Note: The server also honors the conventional `PORT` environment variable. If set, it overrides the configured port when binding the HTTP listener.

## Running locally

From the repository root:

```bash
go run ./cmd/game-room-manager
```

Override the port and environment via env vars:

```bash
APP_PORT=9090 APP_ENV=dev APP_LOG_LEVEL=debug go run ./cmd/game-room-manager
```

The service exposes:

- `GET /healthz` – simple health check
- `GET /readyz` – simple readiness check

## Building and running with Docker

Build the image:

```bash
docker build -t game-room-manager .
```

Run the container, exposing port 8080:

```bash
docker run --rm -p 8080:8080 game-room-manager
```

Override the HTTP port:

```bash
docker run --rm -p 9090:9090 -e APP_PORT=9090 game-room-manager
```

## Running with Docker Compose

A docker-compose.yml file is provided to run the complete environment, which includes:

- game-room-manager: This service (exposed on port 8080).
- planbackend: The planning API service (exposed on port 5000). It assumes the planning backend repository shares the same root as this repo.
- redis: Redis cache used by the planning API (exposed on port 6379).

To start all services in the background, run:

`bash docker compose up -d --build`

To stop the services, run:

`bash docker compose down`

## Logging

Logging uses Zap:

- In `local` / `dev` envs, it uses Zap's development config (human-friendly).
- In other envs, it uses Zap's production config (JSON, structured).
- Log level is controlled by `log_level` / `APP_LOG_LEVEL`.

On startup, the service logs a summary of the active configuration and environment.

## Testing

Run all tests:

```bash
go test ./...
```

## Setting Environment

This projeect uses environment variables to determine certain functions.

For example depending on the platform the command used to launch game rooms differs.

| ENV VARIABLE  | DESCRIPTION                               | EXAMPLE                                                 |
| ------------- | ----------------------------------------- | ------------------------------------------------------- |
| GAME_ROOM_CMD | Sets the command used to launch game room | ./gameroom.app/Contents/MacOS/Core --headless -- --port |
| IP_ADRESS     | Sets the ip address of created rooms      | 127.0.0.1                                               |
