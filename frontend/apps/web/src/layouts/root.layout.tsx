import { useState } from "react";
import type { JSX } from "react";
import {
  Outlet,
  useLocation,
  useRouter,
  useSearch,
} from "@tanstack/react-router";
import type { ParsedLocation } from "@tanstack/react-router";
import { Toaster } from "@pdv/ui-kit/components/sonner";
import { Tooltip } from "@pdv/ui-kit/components/tooltip";
import { useIsMobile } from "@pdv/ui-kit/hooks/use-is-mobile.hook";
import { Header } from "../components/header";
import { Sidebar } from "../components/sidebar";
import {
  operationsNav,
  primaryNav,
} from "../components/sidebar/navigation.config";
import type { Optional } from "@pdv/types";
import type { NavItem } from "../interfaces/app-shell.interface";

export function RootLayout(): JSX.Element {
  const isMobile: boolean = useIsMobile();
  const router = useRouter();
  const location: ParsedLocation<object> = useLocation();
  const search = useSearch({ from: "__root__" });
  const isMobileOpen: boolean = search.sidebar_mobile === true;
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState<boolean>(false);

  const current: Optional<NavItem> = [...primaryNav, ...operationsNav].find(
    (item: NavItem): boolean => item.to === location.pathname,
  );

  function updateUrl(params: Record<string, string | undefined>): void {
    const nextUrl: URL = new URL(window.location.href);

    for (const [key, value] of Object.entries(params)) {
      if (value === undefined) {
        nextUrl.searchParams.delete(key);
      } else {
        nextUrl.searchParams.set(key, value);
      }
    }

    router.history.push(`${location.pathname}${nextUrl.search}${nextUrl.hash}`);
  }

  function closeMobileMenu(): void {
    updateUrl({ sidebar_mobile: undefined });
  }

  function openMenu(): void {
    if (isMobile) {
      updateUrl({ sidebar_mobile: "true" });
    }
  }

  function toggleSidebar(): void {
    if (isMobile) {
      updateUrl({ sidebar_mobile: isMobileOpen ? undefined : "true" });
      return;
    }

    setIsSidebarCollapsed((isCollapsed: boolean) => !isCollapsed);
  }

  return (
    <Tooltip.Provider>
      <div className="min-h-svh bg-(--paper)">
        <div className="app-grain fixed inset-0 z-0" />
        <Sidebar
          isMobileOpen={isMobileOpen}
          isCollapsed={isSidebarCollapsed}
          onClose={closeMobileMenu}
        />
        <button
          type="button"
          aria-label="Fechar menu"
          onClick={closeMobileMenu}
          className={`fixed inset-0 z-30 bg-(--ink)/45 backdrop-blur-sm motion-safe:transition-opacity motion-safe:duration-300 md:hidden ${isMobileOpen ? "opacity-100" : "pointer-events-none opacity-0"}`}
        />
        <main
          className={`relative z-10 min-h-svh motion-safe:transition-[padding] motion-safe:duration-300 motion-safe:ease-[cubic-bezier(0.22,1,0.36,1)] ${isSidebarCollapsed ? "md:pl-18" : "md:pl-66.5"}`}
        >
          <Header
            current={current}
            isMobile={isMobile}
            isCollapsed={isSidebarCollapsed}
            onOpenMenu={openMenu}
            onNavigate={closeMobileMenu}
            onToggleSidebar={toggleSidebar}
          />
          <div className="mx-auto w-full max-w-360 px-5 py-8 sm:px-8 lg:px-10 lg:py-10">
            <div
              key={location.pathname}
              className="motion-safe:animate-page-enter"
            >
              <Outlet />
            </div>
          </div>
        </main>
        <Toaster />
      </div>
    </Tooltip.Provider>
  );
}
