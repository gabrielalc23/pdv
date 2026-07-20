import type {
  ApiErrorLike,
  AuthTransportConfiguration,
} from "../types/transport.types";

let currentConfig: AuthTransportConfiguration | null = null;

export function configureAuthTransport(
  config: AuthTransportConfiguration,
): () => void {
  currentConfig = config;
  return (): void => {
    if (currentConfig === config) {
      currentConfig = null;
    }
  };
}

export function resetAuthTransportConfiguration(): void {
  currentConfig = null;
}

export function getAuthTransportConfiguration(): AuthTransportConfiguration | null {
  return currentConfig;
}

export function isRefreshEligibleError(error: ApiErrorLike): boolean {
  const eligibleCodes = [
    "ACCESS_TOKEN_EXPIRED",
    "AUTHORIZATION_STALE",
    "AUTH_CONTEXT_STALE",
  ];
  return (
    error.status === 401 &&
    error.code !== null &&
    eligibleCodes.includes(error.code)
  );
}

export function isTerminalAuthError(error: ApiErrorLike): boolean {
  const terminalCodes = [
    "SESSION_REVOKED",
    "SESSION_EXPIRED",
    "REFRESH_TOKEN_REUSED",
    "REFRESH_TOKEN_INVALID",
    "REFRESH_TOKEN_EXPIRED",
  ];
  return terminalCodes.includes(error.code ?? "");
}

export function isExpectedAnonymousRefreshError(error: ApiErrorLike): boolean {
  return error.code === "REFRESH_TOKEN_MISSING";
}
