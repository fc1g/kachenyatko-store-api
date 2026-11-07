FROM golang:1.27.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server ./cmd/server
COPY internal ./internal
COPY docs ./docs

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM golang:1.27.1 AS build-create-admin

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/create-admin ./cmd/create-admin
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /create-admin ./cmd/create-admin

FROM golang:1.27.1 AS goose
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM gcr.io/distroless/static-debian12 AS migrate

COPY --from=goose /go/bin/goose /goose
COPY db/migrations /migrations

ENV GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=/migrations

USER nonroot:nonroot

ENTRYPOINT [ "/goose", "up" ]

FROM gcr.io/distroless/static-debian12 AS build-release

WORKDIR /

COPY --from=build /server /server

ENV HTTP_HOST=0.0.0.0

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/server" ]

FROM gcr.io/distroless/static-debian12 AS create-admin

COPY --from=build-create-admin /create-admin /create-admin

USER nonroot:nonroot

ENTRYPOINT [ "/create-admin" ]
