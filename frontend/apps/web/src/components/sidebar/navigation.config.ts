import {
  Box,
  Grid2X2,
  LayoutDashboard,
  Package,
  Receipt,
  ShoppingBag,
} from "lucide-react";
import type { NavItem } from "../../interfaces/app-shell.interface";

export const primaryNav: NavItem[] = [
  { label: "Visão geral", to: "/", icon: LayoutDashboard },
  { label: "Novo atendimento", to: "/pos", icon: ShoppingBag },
];

export const operationsNav: NavItem[] = [
  { label: "Vendas", to: "/sales", icon: Receipt },
  { label: "Produtos", to: "/products", icon: Package },
  { label: "Estoque", to: "/inventory", icon: Box },
  { label: "Catálogo", to: "/catalog", icon: Grid2X2 },
];
