# @pdv/web

The customer-facing PDV web application. It is a React 19 and Vite application
using TanStack Router, Tailwind CSS, and the shared `@pdv/ui-kit` package.

## Run

Install workspace dependencies from `frontend` first, then run:

```sh
bun --cwd apps/web run dev
```

During development, requests use `/api` and Vite proxies them to the backend at
`http://localhost:8080`. To use another API URL, set `VITE_API_URL` before starting
the app.

## Commands

```sh
bun --cwd apps/web run dev
bun --cwd apps/web run build
bun --cwd apps/web run preview
bun --cwd apps/web run lint
bun --cwd apps/web run format
bun --cwd apps/web run format:check
```

## Structure

- `src/routes` contains TanStack file-based routes.
- `src/routes/index.tsx` is the current home route.
- `src/queries` and `src/mutations` contain the TanStack Query integration with the API.
- `src/styles` contains the application stylesheet and imports shared UI styles.

## Shared Packages

The app uses `@pdv/ui-kit` for reusable components. Import its stylesheet once
from the global stylesheet:

```ts
import "@pdv/ui-kit/styles.css";
```

The app uses `@pdv/http` for validated API calls and mounts a shared
`QueryClientProvider` in `src/main.tsx`.
