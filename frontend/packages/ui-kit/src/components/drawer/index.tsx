"use client"

import { DrawerClose } from "./drawer-close"
import { DrawerContent } from "./drawer-content"
import { DrawerDescription } from "./drawer-description"
import { DrawerFooter } from "./drawer-footer"
import { DrawerHeader } from "./drawer-header"
import { DrawerOverlay } from "./drawer-overlay"
import { DrawerPortal } from "./drawer-portal"
import { DrawerRoot } from "./drawer-root"
import { DrawerSwipeHandle } from "./drawer-swipe-handle"
import { DrawerTitle } from "./drawer-title"
import { DrawerTrigger } from "./drawer-trigger"

export type DrawerComponentType = typeof DrawerRoot & {
  Close: typeof DrawerClose
  Content: typeof DrawerContent
  Description: typeof DrawerDescription
  Footer: typeof DrawerFooter
  Header: typeof DrawerHeader
  Overlay: typeof DrawerOverlay
  Portal: typeof DrawerPortal
  SwipeHandle: typeof DrawerSwipeHandle
  Title: typeof DrawerTitle
  Trigger: typeof DrawerTrigger
}

export const Drawer: DrawerComponentType = Object.assign(DrawerRoot, {
  Close: DrawerClose,
  Content: DrawerContent,
  Description: DrawerDescription,
  Footer: DrawerFooter,
  Header: DrawerHeader,
  Overlay: DrawerOverlay,
  Portal: DrawerPortal,
  SwipeHandle: DrawerSwipeHandle,
  Title: DrawerTitle,
  Trigger: DrawerTrigger,
}) satisfies DrawerComponentType
