let currentCsrfToken: string | null = null;
const listeners = new Set<(token: string | null) => void>();

export function getCsrfToken(): string | null {
  return currentCsrfToken;
}

export function setCsrfToken(token: string): void {
  currentCsrfToken = token;
  notify();
}

export function clearCsrfToken(): void {
  currentCsrfToken = null;
  notify();
}

export function subscribeToCsrfToken(
  listener: (token: string | null) => void,
): () => void {
  listeners.add(listener);
  return (): void => {
    listeners.delete(listener);
  };
}

function notify(): void {
  for (const listener of Array.from(listeners)) {
    try {
      listener(currentCsrfToken);
    } catch {
      // subscriber error must not corrupt state
    }
  }
}

export function resetCsrfTokenStore(): void {
  currentCsrfToken = null;
  listeners.clear();
}
