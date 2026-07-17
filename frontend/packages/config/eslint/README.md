# @pdv/eslint

Shared ESLint configuration factory for PDV TypeScript workspaces.

## Usage

```ts
import { createWorkspaceEslintConfig } from "@pdv/eslint";

export default createWorkspaceEslintConfig({
  tsconfigRootDir: import.meta.dirname,
});
```

The factory composes the TanStack ESLint configuration, configures type-aware
parser settings, and applies the workspace style rules.

Pass `tsconfigProject` when an application uses different TypeScript project
files.

## Formatting

```sh
bun --cwd packages/config/eslint run format:check
```
