import type { LucideIcon } from "lucide-react";
import type { AppRoute } from "../types/app-shell.type";

export interface NavItem {
  label: string;
  to: AppRoute;
  icon: LucideIcon;
}

export interface HeaderProps {
  current?: NavItem;
  isMobile: boolean;
  isCollapsed: boolean;
  onOpenMenu: VoidFunction;
  onNavigate: VoidFunction;
  onToggleSidebar: VoidFunction;
}

export interface SidebarProps {
  isMobileOpen: boolean;
  isCollapsed: boolean;
  onClose: VoidFunction;
}

export interface NavigationProps {
  isCollapsed: boolean;
  onNavigate: VoidFunction;
}

export interface NavGroupProps {
  label: string;
  items: NavItem[];
  isCollapsed: boolean;
  onNavigate: VoidFunction;
}
