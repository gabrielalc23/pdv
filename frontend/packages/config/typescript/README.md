# @pdv/typescript

Shared TypeScript configuration presets for workspace packages and applications.

## Presets

| Export                       | Use case                                       |
| ---------------------------- | ---------------------------------------------- |
| `@pdv/typescript/base.json`  | Common strict compiler options.                |
| `@pdv/typescript/react.json` | React and Vite applications.                   |
| `@pdv/typescript/node.json`  | Node-oriented configuration files and tooling. |

Example:

```json
{
  "extends": "@pdv/typescript/react.json",
  "include": ["src"]
}
```

The package requires TypeScript 6 as a peer dependency.

## Formatting

```sh
bun --cwd packages/config/typescript run format:check
```
