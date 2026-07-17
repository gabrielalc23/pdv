import type { LucideIcon } from "lucide-react";
import type { MetricTone } from "../types/metric-tone.type";

export interface MetricProps {
  label: string;
  value: string;
  change: string;
  tone: MetricTone;
  icon: LucideIcon;
}
