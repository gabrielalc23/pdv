import { Link } from "@tanstack/react-router";
import type { LucideIcon } from "lucide-react";
import type { JSX } from "react";
import type { NavGroupProps } from "../../interfaces/app-shell.interface";

export function NavGroup({
  label,
  items,
  isCollapsed,
  onNavigate,
}: NavGroupProps): JSX.Element {
  return (
    <div>
      <p
        className={`mb-2 max-h-8 overflow-hidden px-3 text-[12px] font-semibold tracking-[0.2em] text-(--contrast-light)/35 motion-safe:transition-[max-height,margin,opacity] motion-safe:duration-200 motion-safe:ease-out ${isCollapsed ? "md:mb-0 md:max-h-0 md:opacity-0" : "md:mb-2 md:max-h-8 md:opacity-100"}`}
      >
        {label}
      </p>
      <div className="space-y-1">
        {items.map((item) => {
          const Icon: LucideIcon = item.icon;

          return (
            <Link
              key={item.to}
              to={item.to}
              onClick={onNavigate}
              activeOptions={{ exact: true, includeSearch: false }}
              activeProps={{
                className:
                  "bg-(--surface-muted) text-(--ink) shadow-(--shadow-active) md:translate-x-1",
              }}
              inactiveProps={{
                className:
                  "translate-x-0 text-(--contrast-light)/60 hover:bg-(--contrast-light)/10 hover:text-(--contrast-light)",
              }}
              className={`group flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium motion-safe:transition-[background-color,color,box-shadow,transform,gap] motion-safe:duration-300 motion-safe:ease-out ${isCollapsed ? "md:w-10 md:justify-center md:gap-0 md:px-0" : ""}`}
              title={isCollapsed ? item.label : undefined}
            >
              <Icon className="size-4 shrink-0 opacity-80 transition-transform group-hover:scale-105" />
              <span
                className={`max-w-40 overflow-hidden whitespace-nowrap motion-safe:transition-[max-width,opacity,transform] motion-safe:duration-200 motion-safe:ease-out ${isCollapsed ? "md:max-w-0 md:-translate-x-2 md:opacity-0" : "md:max-w-40 md:translate-x-0 md:opacity-100"}`}
              >
                {item.label}
              </span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
