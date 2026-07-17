import type { JSX } from "react";
import type { ActivityRowConfig } from "../../interfaces/activity-row-config.interface";
import type { ActivityRowProps } from "../../interfaces/activity-row-props.interface";
import { activityRowConfig } from "./activity-row.config";

export function ActivityRow({
  title,
  detail,
  amount,
  time,
  tone,
}: ActivityRowProps): JSX.Element {
  const { icon: Icon, className }: ActivityRowConfig = activityRowConfig[tone];

  return (
    <div className="flex items-center gap-4 px-5 py-4 sm:px-6">
      <span
        className={`grid size-9 shrink-0 place-items-center rounded-md ${className}`}
      >
        <Icon className="size-4.25" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold">{title}</p>
        <p className="mt-0.5 truncate text-xs text-(--ink-soft)">{detail}</p>
      </div>
      <div className="text-right">
        <p className="text-sm font-semibold">{amount}</p>
        <p className="mt-0.5 text-xs text-(--ink-soft)">{time}</p>
      </div>
    </div>
  );
}
