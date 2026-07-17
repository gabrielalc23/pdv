import { Drawer as DrawerPrimitive } from "@base-ui/react/drawer";

import { cn } from "#lib/utils";

export function DrawerDescription({ className, ...props }: DrawerPrimitive.Description.Props) {
  return (
    <DrawerPrimitive.Description
      data-slot="drawer-description"
      className={cn("text-sm text-balance text-muted-foreground", className)}
      {...props}
    />
  );
}
