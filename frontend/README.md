# PDV Frontend

This directory is the Bun workspace for PDV web clients and shared TypeScript
packages.

## Workspace Layout

| Directory | Description |
| --- | --- |
| [`apps/web`](./apps/web) | Customer-facing POS web application. |
| [`apps/admin`](./apps/admin) | Administrative web application. |
| [`packages`](./packages) | Shared UI, HTTP, type, utility, error, and configuration packages. |

## Requirements

- Bun
- Node-compatible toolchain required by Vite and TypeScript.

## Install

Run once from this directory:

```sh
bun install
```

## Common Commands

```sh
# Start the web app
bun run dev:web

# Build every workspace that defines a build script
bun run build

# Build the web app only
bun run build:web

# Check formatting and lint every applicable workspace
bun run format:check
bun run lint
```

Run commands for a specific workspace with Bun's current-working-directory
option. For example:

```sh
bun --cwd apps/admin run dev
bun --cwd packages/http run typecheck
```

## Conventions

- Apps use React, Vite, TanStack Router, and Tailwind CSS.
- Shared packages are consumed through their `@pdv/*` workspace names.
- Formatting is configured by `@pdv/prettier`.
- ESLint configuration is provided by `@pdv/eslint`.
- TypeScript presets are provided by `@pdv/typescript`.

## Applications

- [Web](./apps/web/README.md)
- [Admin](./apps/admin/README.md)
