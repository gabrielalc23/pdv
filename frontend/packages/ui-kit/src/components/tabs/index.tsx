"use client";

import { TabsContent } from "./tabs-content";
import { TabsList } from "./tabs-list";
import { TabsRoot } from "./tabs-root";
import { TabsTrigger } from "./tabs-trigger";

type TabsComponentType = typeof TabsRoot & {
  Content: typeof TabsContent;
  List: typeof TabsList;
  Trigger: typeof TabsTrigger;
};

export const Tabs: TabsComponentType = Object.assign(TabsRoot, {
  Content: TabsContent,
  List: TabsList,
  Trigger: TabsTrigger,
}) satisfies TabsComponentType;
