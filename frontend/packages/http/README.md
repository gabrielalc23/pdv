# @pdv/http

Shared Axios-based HTTP utilities for frontend applications.

## Exports

The package exposes Axios instances, `HttpMethod`, and `createApiCall` for
validated API calls.

```ts
import { createApiCall, instance } from "@pdv/http"
```

The default instance uses `NEXT_PUBLIC_API_URL` as its base URL and falls back
to `https://api.seusite.com`. It sends credentials and attaches the
`accessToken` value from `localStorage` as a bearer token when present.

Applications must provide compatible `axios` and `zod` dependencies, as well as
the `@pdv/errors` and `@pdv/utils` workspace packages.

## Commands

```sh
bun --cwd packages/http run typecheck
bun --cwd packages/http run lint
bun --cwd packages/http run format:check
```
