export const REFRESH_ELIGIBLE_CODES = [
  "ACCESS_TOKEN_EXPIRED",
  "AUTHORIZATION_STALE",
  "AUTH_CONTEXT_STALE",
] as const;

export const TERMINAL_AUTH_CODES = [
  "SESSION_REVOKED",
  "SESSION_EXPIRED",
  "REFRESH_TOKEN_REUSED",
  "REFRESH_TOKEN_INVALID",
  "REFRESH_TOKEN_EXPIRED",
  "REFRESH_TOKEN_MISSING",
] as const;

export type RefreshEligibleCode = (typeof REFRESH_ELIGIBLE_CODES)[number];
export type TerminalAuthCode = (typeof TERMINAL_AUTH_CODES)[number];
