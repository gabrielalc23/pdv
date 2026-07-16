# @pdv/types

Shared TypeScript utility types used by PDV applications and packages.

## Exports

- `Optional<T>`
- `Nullable<T>`
- `Either` types and factories

```ts
import type { Nullable, Optional } from "@pdv/types"
import { left, right } from "@pdv/types"
```

## Commands

```sh
bun --cwd packages/types run typecheck
bun --cwd packages/types run lint
bun --cwd packages/types run format:check
```
