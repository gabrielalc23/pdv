# @pdv/web

The customer-facing PDV web application. It is a React 19 and Vite application
using TanStack Router, Tailwind CSS, and the shared `@pdv/ui-kit` package.

## Run

Install workspace dependencies from `frontend` first, then run:

```sh
bun --cwd apps/web run dev
```

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
- `src/routes/about.tsx` is the current about route.
- `src/styles` contains the application stylesheet and imports shared UI styles.

## Shared Packages

The app uses `@pdv/ui-kit` for reusable components. Import its stylesheet once
from the global stylesheet:

```ts
import "@pdv/ui-kit/styles.css"
```

The app currently does not configure an API client. Add API configuration only
when the frontend-to-backend integration is implemented.
