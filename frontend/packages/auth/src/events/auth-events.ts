import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import type { AuthLossReason } from "../types/auth.types";

export type AuthEvent =
  | { type: "token-updated" }
  | { type: "session-updated"; session: AuthSessionResponse }
  | { type: "auth-lost"; reason: AuthLossReason }
  | { type: "logout" }
  | { type: "context-changed"; session: AuthSessionResponse }
  | { type: "bootstrap-completed"; authenticated: boolean };

const listeners = new Set<(event: AuthEvent) => void>();

export function subscribeToAuthEvents(
  listener: (event: AuthEvent) => void,
): () => void {
  listeners.add(listener);
  return (): void => {
    listeners.delete(listener);
  };
}

export function dispatchAuthEvent(event: AuthEvent): void {
  for (const listener of Array.from(listeners)) {
    try {
      listener(event);
    } catch {
      // subscriber error must not corrupt state
    }
  }
}

export function resetAuthEventBus(): void {
  listeners.clear();
}
