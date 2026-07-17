import { Tooltip as TooltipPrimitive } from "@base-ui/react/tooltip";

export function TooltipRoot({ ...props }: TooltipPrimitive.Root.Props) {
  return <TooltipPrimitive.Root data-slot="tooltip" {...props} />;
}
