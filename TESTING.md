# План тестування

Мінімальний план: лише `service` і `handler` — бізнес-логіка і контракт API.
Усе тут — звичайні юніт-тести, без БД, без R2 і без build-тегів.

Позначки: ✅ готово · 🔄 оновити після змін · ⏳ треба написати · ⏸ після переробки коду

## Інструменти

| Позначка | Що це |
|---|---|
| `testify` | `assert` + `require` |
| `httptest` | `net/http/httptest` — `NewRecorder`, `NewRequest` |
| `chi ctx` | `chi.NewRouteContext()` + `rctx.URLParams.Add(...)` — підкласти `{id}` у запит без роутера |
| `fake` | ручний фейк інтерфейсу (struct з полями-функціями; незадане поле → panic = «не мало викликатись») |
| `validator` | справжній `validation.New()` |
| `pgconn` | `&pgconn.PgError{Code: pgerrcode.X}` і `pgx.ErrNoRows` — імітація помилок БД |
| `bytes` | `bytes.NewReader(...)` — це `io.ReadSeeker`, замінює файл у `ImageFile` |
| `magic` | перші байти файлу для `http.DetectContentType`: PNG `"\x89PNG\r\n\x1a\n"`, JPEG `"\xff\xd8\xff"` |
| `multipart` | `mime/multipart.NewWriter` + `CreateFormFile("image", "cat.png")` — тіло multipart-запиту |

## Дерево

```
toy-shop
├── internal/config
│   └── Load()                                          ✅ (R2_* додані в setValidEnv і want)
│
├── internal/service
│   ├── validate.go         validate(), fieldMessage()  ✅
│   ├── category.go         categoryError()             ✅
│   │                       CategoryService             ✅ (вкл. GetAll з limit/offset)
│   │
│   ├── image.go
│   │   ├── detectImage()                               ⏳ (testify, bytes, magic, table-driven)
│   │   │     ├── PNG / JPEG → contentType + ext, Body перемотано на початок
│   │   │     ├── маленький файл (< 512 байт) → без помилки
│   │   │     ├── Size = 0 → INVALID "image is empty"
│   │   │     ├── Size > 5 MB → INVALID "image must be at most 5 MB"
│   │   │     └── текст → INVALID "image must be JPEG, PNG or WebP"
│   │   └── deleteImages()                              ⏳ (fake imageStorage, -race)
│   │         ├── Delete викликано для кожного ключа (зібрати ключі під mutex)
│   │         └── один Delete впав → решта все одно викликані
│   │
│   ├── toy_image.go
│   │   ├── toyImageError()                             ⏳ (pgconn, table-driven)
│   │   │     ErrNoRows→NOT_FOUND, UniqueViolation→CONFLICT,
│   │   │     ForeignKeyViolation→NOT_FOUND "toy not found", інше → загорнуто
│   │   └── ToyImageService                             ⏳ (fake toyImageStore + fake Querier + fake imageStorage, -race)
│   │       ├── Create
│   │       │     ├── 0 файлів / > 10 файлів → INVALID, нічого не викликається
│   │       │     ├── успіх (3 файли): 3 Upload, 3 insert у InTx, ключі "toys/<toyID>/<uuid>.<ext>"
│   │       │     ├── один файл невалідний → жодного Upload (перевірка ДО завантаження)
│   │       │     ├── один Upload впав → insert не викликається, Delete для ВСІХ ключів
│   │       │     └── insert впав (FK) → NOT_FOUND, Delete для всіх ключів
│   │       ├── GetOne   у репо пішли і toyID, і id
│   │       └── Delete   ключ з репо → storage.Delete; ErrNoRows → NOT_FOUND, storage не викликається
│   │
│   └── toy.go
│       ├── toyError()                                  ⏳ ErrNoRows→NOT_FOUND, ForeignKeyViolation→INVALID
│       └── ToyService                                  ⏳ (fake toyStore: InTx викликає fn з fake Querier)
│             ├── Create   upsert категорії → CreateToy з її ID; невалідний DTO → репо не викликається
│             └── Delete
│                   ├── ключі з DeleteImagesByToyID → storage.Delete для кожного
│                   ├── count=0 → NOT_FOUND, storage не викликається
│                   └── storage.Delete впав → все одно nil (іграшка вже видалена)
│
└── internal/handler
    ├── errors.go           statusFor(), writeError()   ✅ (вкл. CodeTooLarge → 413)
    ├── json.go             writeJSON(), decodeJSON()   ✅
    ├── params.go           parseIDParam()              ✅ (через handler-тести)
    │                       parsePagination()           ✅
    ├── category.go         CategoryHandler             ✅ (вкл. ?limit=&offset=)
    └── toy_image.go
        └── ToyImageHandler.Create                      ⏳ (multipart, chi ctx, fake toyImageService, fake urlBuilder)
              ├── успіх (2 файли в полі "images") → 201, масив, у сервіс прийшли toyID і 2 файли,
              │     url = urls.URL(storage_key)
              ├── не multipart → 400 "invalid multipart form"
              ├── тіло > ліміту (обгорнути r.Body у MaxBytesReader) → 413
              └── помилка сервісу → статус з apperr
```

Не тестуємо окремо:
- `ToyHandler` і `ToyImageHandler.GetAll/GetOne/Delete` — та сама схема, що й у `CategoryHandler`.
- `storage.R2` — тонка обгортка над SDK, перевіряється вручну через Postman.

## Порядок

1. `detectImage` → `toyImageError` → `toyError` — таблиці, без фейків.
2. `deleteImages` — перший тест з горутинами: фейк storage викликається паралельно,
   тому все, що він записує, — під `sync.Mutex`; запуск з `go test -race`.
3. `ToyImageService` — фейк store (`InTx` викликає `fn` з фейковим `Querier`) + фейк storage.
4. `ToyService` — той самий фейк store.
5. `ToyImageHandler.Create` — multipart-запит у тесті.

## Відкладено

`validation`, `logging`, `httpmw`, `router`, `storage`, `repository` (testcontainers), `app`, e2e.









# Kachenyatko store api

## REST API for a small handmade toy shop

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

Check that it works:

curl http://localhost:3000/livez    # {"status":"OK"}
curl http://localhost:3000/readyz   # {"status":"OK"} when the database is reachable

API docs: http://localhost:3000/swagger/index.html
```

### Configuration

The server reads its configuration from environment variables. In development they are loaded from .env if the file exists; in production, set them in the environment directly.

|         Variable          |   Required    |            Default            |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|  ENVIRONMENT              |  yes          |  —                            |
|  PORT                     |  no           |  3000                         |
|  LOG_LEVEL                |  yes          |  —                            |
|  LOG_PRETTY               |  no           |  true in dev, otherwise false |
|  DB_URL                   |  yes          |  —                            |
|  DB_PING_TIMEOUT          |  no           |  2s                           |
|  HTTP_READ_HEADER_TIMEOUT |  no           |  5s                           |
|  HTTP_READ_TIMEOUT        |  no           |  10s                          |
|  HTTP_WRITE_TIMEOUT       |  no           |  10s                          |
|  HTTP_IDLE_TIMEOUT        |  no           |  60s                          |
|  HTTP_SHUTDOWN_TIMEOUT    |  no           |  10s                          |
|  R2_ENDPOINT              |  yes          |  —                            |
|  R2_ACCESS_KEY_ID         |  yes          |  —                            |
|  R2_SECRET_ACCESS_KEY     |  yes          |  —                            |
|  R2_BUCKET_NAME           |  yes          |  —                            |
|  R2_PUBLIC_URL            |  yes          |  —                            |

-----------------------------------------------------------------------------

|         Variable          |                  Description                  |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|  ENVIRONMENT              |  development, production or test              |
|  PORT                     |  HTTP port                                    |
|  LOG_LEVEL                |  trace,debug,info,warn,error,fatal or panic   |
|  LOG_PRETTY               |  Human-readable console logs                  |
|  DB_URL                   |  PostgreSQL connection string                 |
|  DB_PING_TIMEOUT          |  Database ping timeout                        |
|  HTTP_READ_HEADER_TIMEOUT |  Time to read request headers                 |
|  HTTP_READ_TIMEOUT        |  Time to read the whole request               |
|  HTTP_WRITE_TIMEOUT       |  Time to write the response                   |
|  HTTP_IDLE_TIMEOUT        |  Keep-alive idle timeout                      |
|  HTTP_SHUTDOWN_TIMEOUT    |  Graceful shutdown timeout                    |
|  R2_ENDPOINT              |  https://<account-id>.r2.cloudflarestorage.com|
|  R2_ACCESS_KEY_ID         |  R2 API token access key                      |
|  R2_SECRET_ACCESS_KEY     |  R2 API token secret                          |
|  R2_BUCKET_NAME           |  Bucket for toy images                        |
|  R2_PUBLIC_URL            |  Public base URL of the bucket                |

### Make commands

|         Command           |                  Description                  |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
|  `make up`                |  Start PostgreSQL in Docker (foreground)      |
|  `make down`              |  Stop and remove the containers; data is kept |
|  `make clean`             |  Stop containers and **delete volume**        |
|  `make dev`               |  Run the server with hot reload (air)         |
|  `make migrate`           |  Apply all pending migrations                 |
|  `make migrate-down`      |  Roll back the last applied migration         |
|  `make migrate-status`    |  Show which migrations are applied            |
|  `make psql`              |  Open a `psql` shell in the database container|
|  `make sqlc`              |  Regenerate `internal/repository`             |
|  `make swag`              |  Regenerate Swagger docs in `docs/`           |

All commands read variables from `.env`, so create it first (`cp .env.example .env`).
