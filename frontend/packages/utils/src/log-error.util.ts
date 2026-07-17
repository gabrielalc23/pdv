/**
 * Logs an error message to the console, with optional error details.
 *
 * @remarks
 * - Always logs to console, regardless of environment (unlike `log()`, this has no dev/prod branching).
 * - If `error` is an `Error` instance, logs its `message` and `stack` (if present) on separate lines.
 * - If `error` is any other value, logs its string coercion via `String(error)`.
 *
 * @param message - The main error message.
 * @param error - Optional additional error context (exception, unknown thrown value, etc.).
 */
export function logError(message: string, error?: unknown): void {
  if (error === undefined) {
    console.error(`❌ ${message}`);
    return;
  }

  if (error instanceof Error) {
    console.error(`❌ ${message}`, error);
    return;
  }

  console.error(`❌ ${message}`, String(error));
}
