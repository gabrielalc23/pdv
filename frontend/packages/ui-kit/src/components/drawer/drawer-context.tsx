import * as React from "react"
import { Drawer as DrawerPrimitive } from "@base-ui/react/drawer"

export type DrawerContextProps = {
  hasSnapPoints: boolean
  modal: DrawerPrimitive.Root.Props["modal"]
  showSwipeHandle: boolean
  swipeDirection: NonNullable<DrawerPrimitive.Root.Props["swipeDirection"]>
}

export const DrawerContext = React.createContext<DrawerContextProps | null>(null)
