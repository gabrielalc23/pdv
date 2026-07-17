import { createFileRoute } from "@tanstack/react-router";
import type { JSX } from "react";

export const Route = createFileRoute("/")({
  component: AdminHome,
});

function AdminHome(): JSX.Element {
  return <main className="p-2">Admin</main>;
}
