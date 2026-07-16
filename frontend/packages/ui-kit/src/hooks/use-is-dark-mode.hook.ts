import { useIsMediaQuery } from "./use-is-media-query.hook"

export function useIsDarkMode(): boolean {
  return useIsMediaQuery("(prefers-color-scheme: dark)")
}
