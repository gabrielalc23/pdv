import type { ActivityRowTone } from "../types/activity-row-tone.type";

export interface ActivityRowProps {
  title: string;
  detail: string;
  amount: string;
  time: string;
  tone: ActivityRowTone;
}
