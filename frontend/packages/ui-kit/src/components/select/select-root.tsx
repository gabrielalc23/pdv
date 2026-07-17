import { Select as SelectPrimitive } from "@base-ui/react/select";

export function SelectRoot<TValue, TMultiple extends boolean | undefined = false>(
  props: SelectPrimitive.Root.Props<TValue, TMultiple>,
) {
  return <SelectPrimitive.Root data-slot="select" {...props} />;
}
