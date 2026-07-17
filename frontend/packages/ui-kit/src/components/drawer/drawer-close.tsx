import { Drawer as DrawerPrimitive } from "@base-ui/react/drawer";

export function DrawerClose({ ...props }: DrawerPrimitive.Close.Props) {
  return <DrawerPrimitive.Close data-slot="drawer-close" {...props} />;
}
