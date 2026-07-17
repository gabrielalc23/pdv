import { createFileRoute } from "@tanstack/react-router";
import { PosPage } from "../pages/pos.page";

export const Route = createFileRoute("/pos")({ component: PosPage });
