import { ArrowUpRight } from "lucide-react";
import type { QuickActionProps } from "../../interfaces/quick-actions-props.interface";
import type { JSX } from "react";

export function QuickAction({
  icon: Icon,
  title,
  description,
}: QuickActionProps): JSX.Element {
  return (
    <button
      type="button"
      className="group flex items-center gap-4 rounded-md border border-(--line) bg-(--contrast-light)/60 p-4 text-left transition-all hover:-translate-y-0.5 hover:bg-(--contrast-light)"
    >
      <span className="grid size-10 place-items-center rounded-md bg-(--blue-wash) text-(--blue-foreground)">
        <Icon className="size-4.5" />
      </span>

      <span>
        <strong className="block text-sm">{title}</strong>

        <small className="mt-1 block text-xs text-(--ink-soft)">
          {description}
        </small>
      </span>

      <ArrowUpRight className="ml-auto size-4 text-(--neutral-dark)/25" />
    </button>
  );
}
