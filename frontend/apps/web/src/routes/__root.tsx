import { createRootRoute, Link, Outlet } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { Toaster } from "@pdv/ui-kit/components/sonner"
import { Tooltip } from "@pdv/ui-kit/components/tooltip"

const RootLayout = () => (
  <Tooltip.Provider>
    <div className="p-2 flex gap-2">
      <Link to="/" className="[&.active]:font-bold">
        Home
      </Link>{" "}
      <Link to="/about" className="[&.active]:font-bold">
        About
      </Link>
    </div>
    <hr />
    <Outlet />
    <TanStackRouterDevtools />
    <Toaster />
  </Tooltip.Provider>
)

export const Route = createRootRoute({ component: RootLayout })
