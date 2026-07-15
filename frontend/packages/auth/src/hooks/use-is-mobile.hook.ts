import { useIsMediaQuery } from "./use-is-media-query.hook";

const MOBILE_BREAKPOINT: number = 768;

export function useIsMobile(): boolean {
  return useIsMediaQuery(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
}
