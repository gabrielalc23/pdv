/**
 * Logs a message: `console.log` in development, `sendBeacon` to `/api/log` in production.
 *
 * @remarks
 * - Fails silently if `sendBeacon` is unsupported or `/api/log` isn't implemented — no error is thrown either way.
 * - The `sendBeacon` return value (whether the request was queued) is not checked.
 * - No-op in production if `navigator` is unavailable (e.g. server-side).
 *
 * @param message - The message to log.
 */
export function log(message: string): void {
  if (process.env.NODE_ENV === "development") {
    console.log(message);
    return;
  }

  if (typeof navigator !== "undefined") {
    navigator.sendBeacon("/api/log", message);
  }
}
