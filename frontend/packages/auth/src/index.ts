// Runtime
export { createAuthRuntime } from "./runtime/auth-runtime";
export type { AuthRuntime, AuthRuntimeOptions } from "./runtime/auth-runtime";

export {
  configureAuthTransport,
  resetAuthTransportConfiguration,
  isRefreshEligibleError,
  isTerminalAuthError,
  isExpectedAnonymousRefreshError,
} from "./runtime/configure-http";

// Stores
export {
  getAccessToken,
  setAccessToken,
  clearAccessToken,
  subscribeToAccessToken,
  getAccessTokenState,
} from "./stores/access-token.store";
export type { AccessTokenState, Clock } from "./stores/access-token.store";

export {
  getCsrfToken,
  setCsrfToken,
  clearCsrfToken,
  subscribeToCsrfToken,
} from "./stores/csrf-token.store";

export {
  getAuthSessionState,
  setAuthenticatedSession,
  setAnonymousSession,
  resetAuthSession,
  subscribeToAuthSession,
} from "./stores/auth-session.store";
export type { AuthSessionState } from "./stores/auth-session.store";

// Schemas
export {
  AuthUserSchema,
  AuthSessionSchema,
  AuthContextSchema,
  AuthSessionResponseSchema,
  CsrfResponseSchema,
} from "./schemas";
export type {
  AuthUser,
  AuthSession,
  AuthContext,
  AuthSessionResponse,
  CsrfResponse,
} from "./schemas";

// Types
export type {
  AuthBootstrapResult,
  AuthLossReason,
  AuthStatus,
  Scope,
  ApiErrorLike,
  AuthTransportConfiguration,
} from "./types";

export { SCOPE_VALUES } from "./types";

// Scopes
export { hasScope, hasAllScopes, hasAnyScope } from "./scopes/scope.helpers";

// Events
export { subscribeToAuthEvents } from "./events/auth-events";
export type { AuthEvent } from "./events/auth-events";

// Client
export {
  fetchCsrfToken,
  refreshSession,
  logout,
  applyContextChange,
  clearLocalAuthState,
} from "./client";

// Coordinator
export type {
  RefreshCoordinator,
  LockAdapter,
  BroadcastAdapter,
} from "./coordinator";

// Bootstrap
export { bootstrapAuth } from "./runtime/bootstrap";
export type { BootstrapOptions } from "./runtime/bootstrap";
