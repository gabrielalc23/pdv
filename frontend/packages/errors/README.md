# @pdv/errors

Shared application error classes for PDV frontend code.

## Exports

The root export provides `AppError` and specialized errors for conflict,
forbidden, invalid API response, not found, unauthorized, and validation cases.

```ts
import { NotFoundError, ValidationError } from "@pdv/errors";
```

## Commands

Run these commands from `frontend`:

```sh
bun --cwd packages/errors run typecheck
bun --cwd packages/errors run lint
bun --cwd packages/errors run format:check
```

This is an internal workspace package and is not published independently.
