import type { ReactNode } from "react";
import type { QueryClient } from "@tanstack/react-query";

export interface TestWrapperProps {
  children: ReactNode;
  client?: QueryClient;
}
