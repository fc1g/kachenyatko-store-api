# Kachenyatko Store API

REST API for a small handmade toy shop.

### Stack

- **Go** with [chi](https://github.com/go-chi/chi) router
- **PostgreSQL** via [pgx](https://github.com/jackc/pgx); queries generated with [sqlc](https://sqlc.dev), migrations with [goose](https://github.com/pressly/goose)
- **Cloudflare R2** (S3-compatible) for toy images, via AWS SDK v2
- [validator](https://github.com/go-playground/validator) for request validation, [zerolog](https://github.com/rs/zerolog) for logging
- **Swagger / OpenAPI** docs generated with [swag](https://github.com/swaggo/swag)

### Prerequisites

- [Go](https://go.dev/dl/) 1.27.1+ (see `go.mod`)
- [Docker](https://docs.docker.com/get-docker/) with Compose v2 (`docker compose`), used to run PostgreSQL locally
- `make`
- [goose](https://github.com/pressly/goose) v3 for database migrations
- [sqlc](https://sqlc.dev) v1.31+ to generate the repository layer from SQL (needed only when you change `db/queries` or `db/migrations`)
- [air](https://github.com/air-verse/air) for hot reload in development (optional: `go run ./cmd/server` also works)
- A Cloudflare R2 bucket with an API token (Object Read & Write), required for image uploads

### Quick start

```sh
# 1. Clone and enter the project
git clone https://github.com/fc1g/kachenyatko-store-api.git
cd kachenyatko-store-api

# 2. Create a local config from the example
cp .env.example .env

# 3. Start PostgreSQL (keep this terminal open, or use `docker compose up -d`)
make up

# 4. In a new terminal: apply database migrations
make migrate

# 5. Run the server with hot reload (requires air)
make dev
# or without air:
go run ./cmd/server
```

Check that it works:

```sh
curl http://localhost:3000/livez    # {"status":"OK"}
curl http://localhost:3000/readyz   # {"status":"OK"} when the database is reachable
```

API docs: http://localhost:3000/swagger/index.html

### Configuration

The server reads its configuration from environment variables. In development they are loaded from `.env` if the file exists; in production, set them in the environment directly.

| Variable                   | Required | Default                       | Description                                                   |
|----------------------------|----------|-------------------------------|---------------------------------------------------------------|
| `ENVIRONMENT`              | yes      | —                             | `development`, `production` or `test`                         |
| `PORT`                     | no       | `3000`                        | HTTP port                                                     |
| `LOG_LEVEL`                | yes      | —                             | `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `panic` |
| `LOG_PRETTY`               | no       | `true` in dev, otherwise `false` | Human-readable console logs instead of JSON                |
| `DB_URL`                   | yes      | —                             | PostgreSQL connection string                                  |
| `DB_PING_TIMEOUT`          | no       | `2s`                          | Database ping timeout                                         |
| `HTTP_READ_HEADER_TIMEOUT` | no       | `5s`                          | Time to read request headers                                  |
| `HTTP_READ_TIMEOUT`        | no       | `10s`                         | Time to read the whole request                                |
| `HTTP_WRITE_TIMEOUT`       | no       | `10s`                         | Time to write the response                                    |
| `HTTP_IDLE_TIMEOUT`        | no       | `60s`                         | Keep-alive idle timeout                                       |
| `HTTP_SHUTDOWN_TIMEOUT`    | no       | `10s`                         | Graceful shutdown timeout                                     |
| `R2_ENDPOINT`              | yes      | —                             | `https://<account-id>.r2.cloudflarestorage.com`               |
| `R2_ACCESS_KEY_ID`         | yes      | —                             | R2 API token access key                                       |
| `R2_SECRET_ACCESS_KEY`     | yes      | —                             | R2 API token secret                                           |
| `R2_BUCKET_NAME`           | yes      | —                             | Bucket for toy images                                         |
| `R2_PUBLIC_URL`            | yes      | —                             | Public base URL of the bucket                                 |

### Make commands

| Command               | Description                                    |
|-----------------------|------------------------------------------------|
| `make up`             | Start PostgreSQL in Docker (foreground)        |
| `make down`           | Stop and remove the containers; data is kept   |
| `make clean`          | Stop containers and **delete volume**          |
| `make dev`            | Run the server with hot reload (air)           |
| `make migrate`        | Apply all pending migrations                   |
| `make migrate-down`   | Roll back the last applied migration           |
| `make migrate-status` | Show which migrations are applied              |
| `make psql`           | Open a `psql` shell in the database container  |
| `make sqlc`           | Regenerate `internal/repository`               |
| `make swag`           | Regenerate Swagger docs in `docs/`             |

All commands read variables from `.env`, so create it first (`cp .env.example .env`).
