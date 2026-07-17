# PDV Backend

The PDV backend is a Go HTTP API backed by PostgreSQL. It uses Chi for routing,
pgx/v5 for database access, and sqlc for type-safe query generation.

## Capabilities

- Product management and catalog lookup.
- Product categories with filtering across products and catalog.
- Inventory entries, adjustments, balances, and movements.
- Open sales with snapshot sale items and totals.
- Transactional checkout with local payment approval and atomic stock decrements.
- Mock fiscal authorization and structured JSON receipts.

Real payment processors, TEF, real PIX, and SEFAZ integrations are intentionally
out of scope.

## Requirements

- Go 1.26.5
- PostgreSQL 18
- `sqlc`
- Docker and Docker Compose are optional, but recommended for local services.

## Configuration

Copy `.env.example` or export the required variables before running the API.

| Variable       | Required | Default | Description                      |
| -------------- | -------- | ------- | -------------------------------- |
| `DATABASE_URL` | Yes      | -       | PostgreSQL connection string.    |
| `HTTP_ADDRESS` | No       | `:8080` | Address used by the HTTP server. |

Example:

```sh
export DATABASE_URL='postgres://pdv:pdv@localhost:5432/pdv?sslmode=disable'
export HTTP_ADDRESS=':8080'
```

## Run Locally

Start only the API when PostgreSQL is already available:

```sh
go run ./cmd/api
```

Or start the local service stack:

```sh
docker compose up
```

Docker Compose exposes the API on `http://localhost:8080`, PostgreSQL on port
`5432`, and Valkey on port `6379`.

## Database Code

- SQL migrations live in [`migrations`](./migrations).
- sqlc query definitions live in [`queries`](./queries).
- Generated code lives in [`internal/platform/database`](./internal/platform/database).

Do not manually edit generated sqlc files. Regenerate them after changing a query
or migration schema:

```sh
sqlc generate
sqlc vet
```

Apply migrations with the PostgreSQL migration workflow used by your environment
before starting the API.

## Validation

```sh
sqlc generate
sqlc vet
go fmt ./...
go vet ./...
go test -count=1 ./...
```

### End-to-end tests

The e2e suite (`tests/e2e`) boots the real HTTP server against a throwaway
PostgreSQL database, runs the migrations, seeds the payment methods, and
exercises the full API surface through HTTP.

It needs a PostgreSQL reachable at `DATABASE_URL`
(default `postgres://pdv:pdv@localhost:5432/pdv?sslmode=disable`). If no
database is reachable, the suite automatically starts one via
`docker compose -f docker-compose.test.yml up -d`.

```sh
# from the backend directory
make test-e2e
# or
go test -count=1 ./tests/e2e/...
```

A GitHub Actions workflow (`.github/workflows/ci.yml`) runs unit and e2e tests
on every push/PR using a Postgres service container.

## API Overview

All API responses use JSON. Amounts and quantities are represented as strings to
preserve decimal precision.

| Domain    | Endpoints                                                                                                                                                                  |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Health    | `GET /health`                                                                                                                                                              |
| Products  | `POST /products`, `GET /products`, `GET /products/{id}`, `PUT /products/{id}`, `POST /products/{id}/activate`, `POST /products/{id}/deactivate`                            |
| Categories | `POST /categories`, `GET /categories`, `GET /categories/{id}`, `PUT /categories/{id}`, `POST /categories/{id}/activate`, `POST /categories/{id}/deactivate`                 |
| Inventory | `GET /inventory`, `GET /products/{id}/inventory`, `POST /inventory/entries`, `POST /inventory/adjustments`, `GET /products/{id}/inventory/movements`                       |
| Catalog   | `GET /catalog`, `GET /catalog/barcode/{barcode}`, `GET /catalog/{id}`                                                                                                      |
| Sales     | `POST /sales`, `GET /sales`, `GET /sales/{id}`, `POST /sales/{id}/items`, `PUT /sales/{id}/items/{itemId}`, `DELETE /sales/{id}/items/{itemId}`, `POST /sales/{id}/cancel` |
| Checkout  | `POST /sales/{id}/checkout`                                                                                                                                                |
| Payments  | `GET /payment-methods`, `GET /sales/{id}/payments`                                                                                                                         |
| Fiscal    | `GET /sales/{id}/fiscal-document`                                                                                                                                          |
| Receipt   | `GET /sales/{id}/receipt`                                                                                                                                                  |

### Checkout Example

```sh
curl -X POST http://localhost:8080/sales/<sale-id>/checkout \
  -H 'Content-Type: application/json' \
  -d '{
    "payments": [
      {
        "paymentMethodId": "<payment-method-id>",
        "amount": "50.00"
      }
    ]
  }'
```

Checkout is transactional: it validates the open sale, validates payments,
decrements inventory atomically, creates sale movements and approved payments,
completes the sale, and creates a pending fiscal document. Fiscal authorization
is performed by the mock provider after the commercial transaction commits.
