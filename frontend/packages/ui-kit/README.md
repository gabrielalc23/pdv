# @pdv/ui-kit

Componentes compartilhados do `@pdv/web` e `@pdv/admin`, gerados pelo shadcn com primitives do Base UI.

```tsx
import { Button } from "@pdv/ui-kit/components/button";
import { Select } from "@pdv/ui-kit/components/select";

<Select>
  <Select.Trigger>
    <Select.Value placeholder="Selecione uma opcao" />
  </Select.Trigger>
  <Select.Content>
    <Select.Item value="cash">Dinheiro</Select.Item>
  </Select.Content>
</Select>;
```

Os apps devem importar `@pdv/ui-kit/styles.css` uma vez em seu stylesheet global.
