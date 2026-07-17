import type { ReactNode } from "react";

export interface PageHeaderBreadcrumb {
  label: string;
  to?: string;
}

export interface PageHeaderProps {
  breadcrumbs: PageHeaderBreadcrumb[];
  title: ReactNode;
  description: string;
  action?: ReactNode;
  actionPlacement?: "side" | "below";
}
