# @pdv/utils

Shared utility functions for PDV frontend applications.

## Exports

The root export includes class-name composition, logging helpers, and request
location resolution utilities.

```ts
import { cn, resolveRequestLocation } from "@pdv/utils";
```

The package expects compatible `clsx` and `tailwind-merge` peer dependencies.

## Commands

```sh
bun --cwd packages/utils run typecheck
bun --cwd packages/utils run lint
bun --cwd packages/utils run format:check
```
