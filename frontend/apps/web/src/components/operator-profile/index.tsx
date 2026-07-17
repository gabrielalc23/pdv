import { ChevronDown } from "lucide-react";
import type { JSX } from "react";

export function OperatorProfile({
  isCollapsed,
}: {
  isCollapsed: boolean;
}): JSX.Element {
  return (
    <button
      type="button"
      aria-label="Perfil do operador"
      className={`flex h-10 w-full items-center gap-3 rounded-md p-2 text-left text-(--contrast-light) motion-safe:transition-[gap] motion-safe:duration-200 motion-safe:ease-out hover:bg-(--contrast-light)/10 ${isCollapsed ? "md:justify-center md:gap-0" : ""}`}
    >
      <span className="grid size-10 shrink-0 place-items-center rounded-md bg-(--blue-wash) text-xs font-bold text-(--blue-foreground)">
        GC
      </span>
      <span
        className={`min-w-0 max-w-40 flex-1 overflow-hidden whitespace-nowrap leading-tight motion-safe:transition-[max-width,opacity,transform] motion-safe:duration-200 motion-safe:ease-out ${isCollapsed ? "md:max-w-0 md:-translate-x-2 md:opacity-0" : "md:max-w-40 md:translate-x-0 md:opacity-100"}`}
      >
        <span className="block truncate text-sm font-semibold">
          Gabriel Campos
        </span>
        <span className="block truncate text-[11px] text-(--contrast-light)/40">
          Operador · Caixa 01
        </span>
      </span>
      <ChevronDown
        className={`size-4 shrink-0 text-(--contrast-light)/35 motion-safe:transition-[width,height,opacity,transform] motion-safe:duration-200 motion-safe:ease-out ${isCollapsed ? "md:size-0 md:translate-x-2 md:opacity-0" : "md:size-4 md:translate-x-0 md:opacity-100"}`}
      />
    </button>
  );
}
