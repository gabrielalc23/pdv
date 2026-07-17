import { Card } from "@pdv/ui-kit/components/card";
import type { MetricTone } from "../../types/metric-tone.type";
import type { MetricProps } from "../../interfaces/metric-props.interface";
import type { JSX } from "react";

const metricToneClasses: Record<MetricTone, string> = {
  coral: "bg-(--coral-wash) text-(--coral-dark)",
  mint: "bg-(--mint) text-(--mint-foreground)",
  blue: "bg-(--blue-wash) text-(--blue-foreground)",
  sand: "bg-(--sand) text-(--sand-foreground)",
};

export function Metric({
  label,
  value,
  change,
  tone,
  icon: Icon,
}: MetricProps): JSX.Element {
  return (
    <Card className="rounded-2xl border-(--line) bg-(--surface) shadow-none">
      <Card.Content className="p-5">
        <div className="mb-5 flex items-start justify-between">
          <p className="text-xs font-semibold text-(--ink-soft)">{label}</p>

          <span
            className={`grid size-8 place-items-center rounded-md ${metricToneClasses[tone]}`}
          >
            <Icon className="size-4" />
          </span>
        </div>

        <p className="font-serif text-[1.8rem] font-semibold tracking-[-0.04em]">
          {value}
        </p>

        <p className="mt-2 text-xs text-(--mint-foreground)">
          <span className="font-semibold">{change}</span>{" "}
          <span className="text-(--ink-soft)">vs. ontem</span>
        </p>
      </Card.Content>
    </Card>
  );
}
