# Kachenyatko Store API

REST API for a small handmade toy shop.

### Stack

- **Go** with [chi](https://github.com/go-chi/chi) router
- **PostgreSQL** via [pgx](https://github.com/jackc/pgx); queries generated with [sqlc](https://sqlc.dev), migrations with [goose](https://github.com/pressly/goose)
- **Cloudflare R2** (S3-compatible) for toy images, via AWS SDK v2
- **Telegram Bot API** via [go-telegram/bot](https://github.com/go-telegram/bot) for new order notifications
- [rs/cors](https://github.com/rs/cors) for CORS, limited to the storefront origin
- **JWT** (EdDSA / Ed25519) via [golang-jwt](https://github.com/golang-jwt/jwt) in an HttpOnly cookie, passwords hashed with [argon2id](https://github.com/alexedwards/argon2id)
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
- A Telegram bot token from [@BotFather](https://t.me/BotFather) and the ID of the chat that receives order notifications (the bot must be a member of that chat)
- `openssl` to generate the Ed25519 key that signs access tokens

### Quick start

```sh
# 1. Clone and enter the project
git clone https://github.com/fc1g/kachenyatko-store-api.git
cd kachenyatko-store-api

# 2. Create a local config from the example
cp .env.example .env

# 3. Generate the token signing key and put it into .env as AUTH_ACCESS_PRIVATE_KEY
openssl genpkey -algorithm ed25519 -out private.pem
base64 < private.pem | tr -d '\n'

# 4. Start PostgreSQL (keep this terminal open, or use `docker compose up -d`)
make up

# 5. In a new terminal: apply database migrations
make migrate

# 6. Create the admin user (see "Authentication" below)
make admin ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='your-password'

# 7. Run the server with hot reload (requires air)
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
| `TELEGRAM_ACCESS_TOKEN`    | yes      | —                             | Bot token from @BotFather                                     |
| `TELEGRAM_CHAT_ID`         | yes      | —                             | Chat for order notifications (negative for groups)            |
| `TELEGRAM_SEND_TIMEOUT`    | no       | `10s`                         | Timeout for sending a notification                            |
| `SITE_URL`                 | yes      | —                             | Storefront URL: allowed CORS origin and base for toy links    |
| `AUTH_ACCESS_PRIVATE_KEY`  | yes      | —                             | Base64 of the Ed25519 private key in PEM format               |
| `AUTH_ACCESS_ISSUER`       | yes      | —                             | `iss` claim written to and checked in access tokens           |
| `AUTH_ACCESS_TOKEN_TTL`    | no       | `24h`                         | Access token lifetime, at most `24h`                          |

### Authentication

The API has a single user: the **admin**, the owner of the shop. Customers do not sign up: they browse the catalog and place orders without an account, so the storefront stays as simple as possible.

That is why there are only login and logout routes and no registration:

| Route          | Auth | Description                                                       |
|----------------|------|-------------------------------------------------------------------|
| `POST /login`  | no   | Checks email and password, sets the `access_token` cookie         |
| `POST /logout` | no   | Clears the `access_token` cookie, returns `204`                   |
| `GET /me`      | yes  | Returns the logged-in user; the admin panel calls it on page load |

**Public** routes: reading categories, toys and toy images, creating orders, login, logout, health checks and Swagger. **Admin only**: creating, updating and deleting categories, toys and images; reading orders; everything under `/users`.

#### Access token

- A JWT signed with Ed25519, sent as an `HttpOnly`, `SameSite=Lax` cookie (`Secure` in production). The frontend sends requests with `credentials: "include"`; the API and the storefront must be on the same site (e.g. `example.com` and `api.example.com`).
- It lives for **1 day** (`AUTH_ACCESS_TOKEN_TTL`, up to `24h`). There are no refresh tokens: the admin just logs in again. Logout only deletes the cookie, so the token itself stays valid until it expires: the route exists mainly for convenience in development and testing.
- Changing `AUTH_ACCESS_PRIVATE_KEY` invalidates all tokens.

#### Creating the admin

There is no HTTP route that creates users. Create the admin once with the `create-admin` command (it needs only `DB_URL`, `ADMIN_EMAIL` and `ADMIN_PASSWORD`; the password must be 8–128 characters):

```sh
make admin ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='your-password'
```

If a user with this email already exists, the command fails and changes nothing.

#### Room to grow

A single admin is a temporary solution. If the project grows and needs it, roles will be added and customers will get more features.

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
| `make admin`          | Create the admin user (see "Authentication")   |
| `make sqlc`           | Regenerate `internal/repository`               |
| `make swag`           | Regenerate Swagger docs in `docs/`             |

All commands read variables from `.env`, so create it first (`cp .env.example .env`).
