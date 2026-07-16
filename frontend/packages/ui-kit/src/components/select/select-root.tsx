import { Select as SelectPrimitive } from "@base-ui/react/select"

export function SelectRoot<Value, Multiple extends boolean | undefined = false>(
  props: SelectPrimitive.Root.Props<Value, Multiple>,
) {
  return <SelectPrimitive.Root data-slot="select" {...props} />
}
