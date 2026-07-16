import { Drawer as DrawerPrimitive } from "@base-ui/react/drawer"

export function DrawerPortal({ ...props }: DrawerPrimitive.Portal.Props) {
  return <DrawerPrimitive.Portal data-slot="drawer-portal" {...props} />
}
