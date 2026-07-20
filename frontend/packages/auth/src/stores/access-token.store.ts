export interface AccessTokenState {
  token: string | null;
  expiresAt: number | null;
}

export interface Clock {
  now: () => number;
}

const defaultClock: Clock = { now: (): number => Date.now() };

let currentToken: string | null = null;
let currentExpiresAt: number | null = null;
let clock: Clock = defaultClock;
const listeners = new Set<(state: Readonly<AccessTokenState>) => void>();

function notify(): void {
  const snapshot: Readonly<AccessTokenState> = Object.freeze({
    token: currentToken,
    expiresAt: currentExpiresAt,
  });
  for (const listener of Array.from(listeners)) {
    try {
      listener(snapshot);
    } catch {
      // subscriber error must not corrupt state
    }
  }
}

export function setClock(c: Clock): void {
  clock = c;
}

export function getAccessToken(): string | null {
  return currentToken;
}

export function getAccessTokenState(): Readonly<AccessTokenState> {
  return Object.freeze({
    token: currentToken,
    expiresAt: currentExpiresAt,
  });
}

export function setAccessToken(token: string, expiresInSeconds: number): void {
  if (!token || typeof token !== "string" || token.length === 0) {
    throw new Error("Token must be a non-empty string");
  }
  if (
    typeof expiresInSeconds !== "number" ||
    !Number.isFinite(expiresInSeconds) ||
    expiresInSeconds <= 0
  ) {
    throw new Error("expiresInSeconds must be a positive number");
  }
  currentToken = token;
  currentExpiresAt = clock.now() + expiresInSeconds * 1000;
  notify();
}

export function clearAccessToken(): void {
  currentToken = null;
  currentExpiresAt = null;
  notify();
}

export function subscribeToAccessToken(
  listener: (state: Readonly<AccessTokenState>) => void,
): () => void {
  listeners.add(listener);
  return (): void => {
    listeners.delete(listener);
  };
}

export function resetAccessTokenStore(): void {
  currentToken = null;
  currentExpiresAt = null;
  listeners.clear();
  clock = defaultClock;
}
