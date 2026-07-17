import { useIsMediaQuery } from "./use-is-media-query.hook";

const TABLET_BREAKPOINT_MIN = 768;
const TABLET_BREAKPOINT_MAX = 1024;

export function useIsTablet(): boolean {
  return useIsMediaQuery(
    `(min-width: ${TABLET_BREAKPOINT_MIN}px) and (max-width: ${TABLET_BREAKPOINT_MAX - 1}px)`,
  );
}
