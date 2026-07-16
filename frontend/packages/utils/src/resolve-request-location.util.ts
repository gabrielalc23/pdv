export type ApiRequestLocation = "data" | "none" | "params"

export function resolveRequestLocation(
  method: string,
  requestLocation?: ApiRequestLocation,
): ApiRequestLocation {
  if (requestLocation) {
    return requestLocation
  }

  switch (method) {
    case "GET":
    case "HEAD":
      return "params"

    case "OPTIONS":
      return "none"

    default:
      return "data"
  }
}
