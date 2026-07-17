import { createFileRoute } from "@tanstack/react-router";
import { SalesPage } from "../pages/sales.page";

export const Route = createFileRoute("/sales")({ component: SalesPage });
