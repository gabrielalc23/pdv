import { Outlet, createRootRoute } from "@tanstack/react-router";
import { Toaster } from "@pdv/ui-kit/components/sonner";
import { Tooltip } from "@pdv/ui-kit/components/tooltip";
import type { JSX } from "react";

function RootLayout(): JSX.Element {
  return (
    <Tooltip.Provider>
      <Outlet />
      <Toaster />
    </Tooltip.Provider>
  );
}

export const Route = createRootRoute({ component: RootLayout });
