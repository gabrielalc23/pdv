import { createRootRoute } from "@tanstack/react-router";
import { z } from "zod";
import { RootLayout } from "../layouts/root.layout";

export const Route = createRootRoute({
  validateSearch: z.object({
    sheet: z.string().optional(),
    sidebar_mobile: z.coerce.boolean().optional(),
  }),
  component: RootLayout,
});
