import { createFileRoute } from "@tanstack/react-router";
import { InventoryPage } from "../pages/inventory.page";

export const Route = createFileRoute("/inventory")({
  component: InventoryPage,
});
