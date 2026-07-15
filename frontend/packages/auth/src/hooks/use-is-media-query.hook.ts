import { useSyncExternalStore } from "react";

export function useIsMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (cb) => {
      const mql: MediaQueryList = window.matchMedia(query);
      mql.addEventListener("change", cb);
      return () => mql.removeEventListener("change", cb);
    },
    () => window.matchMedia(query).matches,
    () => false,
  );
}
