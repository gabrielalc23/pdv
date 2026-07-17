# @pdv/ui-kit

Reusable React UI primitives shared by `@pdv/web` and `@pdv/admin`. Components
are built with Base UI and styled for the workspace Tailwind setup.

## Usage

This package uses explicit subpath exports rather than a root component export.

```tsx
import { Button } from "@pdv/ui-kit/components/button";
import { Select } from "@pdv/ui-kit/components/select";

export function PaymentMethodSelect() {
  return (
    <Select>
      <Select.Trigger>
        <Select.Value placeholder="Select a payment method" />
      </Select.Trigger>
      <Select.Content>
        <Select.Item value="cash">Cash</Select.Item>
      </Select.Content>
    </Select>
  );
}
```

Import the package stylesheet once in the consuming application's global style
entry point:

```ts
import "@pdv/ui-kit/styles.css";
```

## Available Subpaths

- `@pdv/ui-kit/components/*`
- `@pdv/ui-kit/hooks/*`
- `@pdv/ui-kit/lib/*`
- `@pdv/ui-kit/styles.css`

## Validation

```sh
bun --cwd packages/ui-kit run typecheck
```
