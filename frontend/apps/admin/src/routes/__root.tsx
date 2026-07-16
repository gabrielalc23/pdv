import { Outlet, createRootRoute } from '@tanstack/react-router'
import { Toaster } from "@pdv/ui-kit/components/sonner"
import { Tooltip } from "@pdv/ui-kit/components/tooltip"

function RootLayout() {
  return (
    <Tooltip.Provider>
      <Outlet />
      <Toaster />
    </Tooltip.Provider>
  )
}

export const Route = createRootRoute({ component: RootLayout })
