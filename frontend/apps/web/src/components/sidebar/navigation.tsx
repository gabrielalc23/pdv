import type { JSX } from "react";
import type { NavigationProps } from "../../interfaces/app-shell.interface";
import { NavGroup } from "../nav-group";
import { operationsNav, primaryNav } from "./navigation.config";

export function Navigation({
  isCollapsed,
  onNavigate,
}: NavigationProps): JSX.Element {
  return (
    <nav className="flex flex-1 flex-col gap-5 overflow-x-hidden overflow-y-auto px-3 py-4">
      <NavGroup
        label="Operação"
        items={primaryNav}
        isCollapsed={isCollapsed}
        onNavigate={onNavigate}
      />
      <NavGroup
        label="Gestão"
        items={operationsNav}
        isCollapsed={isCollapsed}
        onNavigate={onNavigate}
      />
    </nav>
  );
}
