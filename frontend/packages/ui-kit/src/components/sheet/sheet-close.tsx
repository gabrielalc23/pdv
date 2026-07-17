import { Dialog as SheetPrimitive } from "@base-ui/react/dialog";

export function SheetClose({ ...props }: SheetPrimitive.Close.Props) {
  return <SheetPrimitive.Close data-slot="sheet-close" {...props} />;
}
