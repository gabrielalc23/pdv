import { Box, Receipt, ShoppingBag } from "lucide-react";
import type { ActivityRowConfig } from "../../interfaces/activity-row-config.interface";
import type { ActivityRowTone } from "../../types/activity-row-tone.type";

export const activityRowConfig: Record<ActivityRowTone, ActivityRowConfig> = {
  coral: {
    icon: Receipt,
    className: "bg-(--coral-wash) text-(--coral-dark)",
  },
  mint: {
    icon: Box,
    className: "bg-(--mint) text-(--mint-foreground)",
  },
  blue: {
    icon: ShoppingBag,
    className: "bg-(--blue-wash) text-(--blue-foreground)",
  },
};
