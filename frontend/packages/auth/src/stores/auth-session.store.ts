import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import type { AuthStatus } from "../types/auth.types";

export interface AuthSessionState {
  status: AuthStatus;
  session: AuthSessionResponse | null;
}

let state: AuthSessionState = {
  status: "unknown",
  session: null,
};

const listeners = new Set<(state: Readonly<AuthSessionState>) => void>();

function notify(): void {
  const snapshot: Readonly<AuthSessionState> = Object.freeze({
    status: state.status,
    session: state.session,
  });
  for (const listener of Array.from(listeners)) {
    try {
      listener(snapshot);
    } catch {
      // subscriber error must not corrupt state
    }
  }
}

export function getAuthSessionState(): Readonly<AuthSessionState> {
  return Object.freeze({
    status: state.status,
    session: state.session,
  });
}

export function setAuthenticatedSession(session: AuthSessionResponse): void {
  state = { status: "authenticated", session };
  notify();
}

export function setAnonymousSession(): void {
  state = { status: "anonymous", session: null };
  notify();
}

export function resetAuthSession(): void {
  state = { status: "unknown", session: null };
  notify();
}

export function subscribeToAuthSession(
  listener: (state: Readonly<AuthSessionState>) => void,
): () => void {
  listeners.add(listener);
  return (): void => {
    listeners.delete(listener);
  };
}

export function resetAuthSessionStore(): void {
  state = { status: "unknown", session: null };
  listeners.clear();
}
