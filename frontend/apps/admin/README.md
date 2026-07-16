# @pdv/admin

The administrative PDV web application. It is a React 19 and Vite application
using TanStack Router, Tailwind CSS, and the shared `@pdv/ui-kit` package.

## Run

Install workspace dependencies from `frontend` first, then run:

```sh
bun --cwd apps/admin run dev
```

## Commands

```sh
bun --cwd apps/admin run dev
bun --cwd apps/admin run build
bun --cwd apps/admin run preview
bun --cwd apps/admin run lint
bun --cwd apps/admin run format
bun --cwd apps/admin run format:check
```

## Structure

- `src/routes` contains TanStack file-based routes.
- `src/routes/index.tsx` is the current admin landing route.
- `src/styles` contains application styles and shared UI imports.
- The root route provides toast and tooltip infrastructure for the app.

## Shared Packages

Use components through their explicit `@pdv/ui-kit` subpath exports and import
`@pdv/ui-kit/styles.css` once in the global stylesheet.

The administrative API integration has not been wired into this app yet.
