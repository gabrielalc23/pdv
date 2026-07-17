"use client";

import { TooltipContent } from "./tooltip-content";
import { TooltipProvider } from "./tooltip-provider";
import { TooltipRoot } from "./tooltip-root";
import { TooltipTrigger } from "./tooltip-trigger";

export type TooltipComponentType = typeof TooltipRoot & {
  Content: typeof TooltipContent;
  Provider: typeof TooltipProvider;
  Trigger: typeof TooltipTrigger;
};

export const Tooltip: TooltipComponentType = Object.assign(TooltipRoot, {
  Content: TooltipContent,
  Provider: TooltipProvider,
  Trigger: TooltipTrigger,
}) satisfies TooltipComponentType;
