import type { JSX } from "react";
import { Link } from "@tanstack/react-router";
import { Store, X } from "lucide-react";
import type { SidebarProps } from "../../interfaces/app-shell.interface";
import { Navigation } from "./navigation";
import { OperatorProfile } from "../operator-profile";

export function Sidebar({
  isMobileOpen,
  isCollapsed,
  onClose,
}: SidebarProps): JSX.Element {
  return (
    <aside
      data-collapsed={isCollapsed}
      className={`fixed inset-0 z-40 flex flex-col overflow-hidden border-r border-(--contrast-light)/10 bg-(--ink) text-(--contrast-light) shadow-2xl motion-safe:transition-all motion-safe:duration-300 motion-safe:ease-[cubic-bezier(0.22,1,0.36,1)] ${isCollapsed ? "md:w-18" : "md:w-66.5"} md:translate-x-0 md:shadow-none ${isMobileOpen ? "translate-x-0 opacity-100" : "-translate-x-full opacity-0"} md:inset-y-0 md:left-0 md:right-auto md:opacity-100`}
    >
      <div className="flex items-center justify-between px-3 pb-4 pt-5">
        <Link
          to="/"
          onClick={onClose}
          className={`group flex min-w-0 items-center gap-3 rounded-md p-2 hover:bg-(--contrast-light)/10 w-full ${isCollapsed ? "md:h-10 md:w-full md:justify-center md:gap-0 md:p-0" : ""}`}
        >
          <span
            className={`grid size-10 shrink-0 place-items-center rounded-md bg-(--coral) text-(--coral-foreground) shadow-(--shadow-coral)`}
          >
            <Store className="size-4" />
          </span>
          <span
            className={`min-w-0 max-w-40 overflow-hidden whitespace-nowrap motion-safe:transition-[max-width,opacity,transform] motion-safe:duration-200 motion-safe:ease-out ${isCollapsed ? "md:max-w-0 md:-translate-x-2 md:opacity-0" : "md:max-w-40 md:translate-x-0 md:opacity-100"}`}
          >
            <strong className="block font-serif text-base font-semibold tracking-[-0.02em]">
              Balcão
            </strong>
            <span className="block truncate text-[10px] uppercase tracking-[0.16em] text-(--contrast-light)/40">
              Sistema de venda
            </span>
          </span>
        </Link>
        <button
          type="button"
          aria-label="Fechar menu"
          onClick={onClose}
          className="rounded-md p-2 text-(--contrast-light)/55 hover:bg-(--contrast-light)/10 md:hidden"
        >
          <X className="size-5" />
        </button>
      </div>
      <Navigation isCollapsed={isCollapsed} onNavigate={onClose} />
      <div className="px-3 py-4">
        <OperatorProfile isCollapsed={isCollapsed} />
      </div>
    </aside>
  );
}
