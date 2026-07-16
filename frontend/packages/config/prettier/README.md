# @pdv/prettier

Shared Prettier configuration for the PDV frontend workspace.

## Usage

```js
import config from "@pdv/prettier";

export default config;
```

The configuration uses LF line endings, a 100-character print width, double
quotes, no semicolons, and trailing commas.

## Formatting

```sh
bun --cwd packages/config/prettier run format:check
```
