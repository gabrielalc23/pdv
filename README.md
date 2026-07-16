# PDV

PDV is a point-of-sale monorepo. It contains a Go HTTP API, a PostgreSQL data
model, and Bun workspaces for the customer-facing and administrative web apps.

## Repository Layout

| Directory | Description |
| --- | --- |
| [`backend`](./backend) | Go API, SQL migrations, sqlc queries, and Docker Compose services. |
| [`frontend`](./frontend) | Bun workspace with web apps, UI components, and shared TypeScript packages. |

## Prerequisites

- Go 1.26.5 for the backend.
- Bun for the frontend workspace.
- PostgreSQL 18, or Docker Compose to run the local services.
- `sqlc` to regenerate and validate database code.

## Quick Start

Start the backend services:

```sh
cd backend
docker compose up
```

In another terminal, install and start the web app:

```sh
cd frontend
bun install
bun run dev:web
```

The API is exposed at `http://localhost:8080` when started with Docker Compose.

## Documentation

- [Backend setup and API overview](./backend/README.md)
- [Frontend workspace guide](./frontend/README.md)
- [Web app](./frontend/apps/web/README.md)
- [Admin app](./frontend/apps/admin/README.md)

## Status

The backend supports products, inventory, catalog search, open sales, checkout,
local payment recording, atomic inventory decrements, a mock fiscal provider,
and JSON receipts. Real payment gateways and fiscal authority integrations are
not part of this repository yet.
