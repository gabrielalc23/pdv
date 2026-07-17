import * as React from "react";

import { useIsMobile } from "#hooks/use-is-mobile.hook";
import { cn } from "#lib/utils";

import {
  SIDEBAR_COOKIE_MAX_AGE,
  SIDEBAR_COOKIE_NAME,
  SIDEBAR_KEYBOARD_SHORTCUT,
  SIDEBAR_WIDTH,
  SIDEBAR_WIDTH_ICON,
} from "./sidebar-constants";
import { SidebarContext } from "./sidebar-context";
import type { SidebarContextProps } from "./sidebar-context";

export default function SidebarProvider({
  defaultOpen: isOpenByDefault = true,
  open: isOpenProp,
  onOpenChange: setOpenProp,
  className,
  style,
  children,
  ...props
}: React.ComponentProps<"div"> & {
  defaultOpen?: boolean;
  open?: boolean;
  onOpenChange?: (isOpen: boolean) => void;
}) {
  const isMobile = useIsMobile();
  const [isOpenMobile, setIsOpenMobile] = React.useState(false);

  // This is the internal state of the sidebar.
  // We use openProp and setOpenProp for control from outside the component.
  const [isInternalOpen, setIsInternalOpen] = React.useState(isOpenByDefault);
  const isOpen = isOpenProp ?? isInternalOpen;
  const setOpen = React.useCallback(
    (value: boolean | ((value: boolean) => boolean)) => {
      const isOpenState = typeof value === "function" ? value(isOpen) : value;
      if (setOpenProp) {
        setOpenProp(isOpenState);
      } else {
        setIsInternalOpen(isOpenState);
      }

      // This sets the cookie to keep the sidebar state.
      document.cookie = `${SIDEBAR_COOKIE_NAME}=${isOpenState}; path=/; max-age=${SIDEBAR_COOKIE_MAX_AGE}`;
    },
    [setOpenProp, isOpen],
  );

  // Helper to toggle the sidebar.
  const toggleSidebar = React.useCallback(() => {
    return isMobile
      ? setIsOpenMobile((isNextOpen) => !isNextOpen)
      : setOpen((isNextOpen) => !isNextOpen);
  }, [isMobile, setOpen, setIsOpenMobile]);

  // Adds a keyboard shortcut to toggle the sidebar.
  React.useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === SIDEBAR_KEYBOARD_SHORTCUT && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        toggleSidebar();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleSidebar]);

  // We add a state so that we can do data-state="expanded" or "collapsed".
  // This makes it easier to style the sidebar with Tailwind classes.
  const state = isOpen ? "expanded" : "collapsed";

  const contextValue = React.useMemo<SidebarContextProps>(
    () => ({
      state,
      open: isOpen,
      setOpen,
      isMobile,
      openMobile: isOpenMobile,
      setOpenMobile: setIsOpenMobile,
      toggleSidebar,
    }),
    [state, isOpen, setOpen, isMobile, isOpenMobile, setIsOpenMobile, toggleSidebar],
  );

  return (
    <SidebarContext.Provider value={contextValue}>
      <div
        data-slot="sidebar-wrapper"
        style={
          {
            "--sidebar-width": SIDEBAR_WIDTH,
            "--sidebar-width-icon": SIDEBAR_WIDTH_ICON,
            ...style,
          } as React.CSSProperties
        }
        className={cn(
          "group/sidebar-wrapper flex min-h-svh w-full has-data-[variant=inset]:bg-sidebar",
          className,
        )}
        {...props}
      >
        {children}
      </div>
    </SidebarContext.Provider>
  );
}
