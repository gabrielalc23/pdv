import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { clearLocalAuthState } from "../client/logout.client";
import {
  resetAccessTokenStore,
  getAccessToken,
  setAccessToken,
} from "../stores/access-token.store";
import {
  resetAuthSessionStore,
  getAuthSessionState,
  setAuthenticatedSession,
} from "../stores/auth-session.store";
import {
  resetCsrfTokenStore,
  getCsrfToken,
  setCsrfToken,
} from "../stores/csrf-token.store";
import { resetAuthEventBus } from "../events/auth-events";
import type { AuthSessionResponse } from "../schemas/auth-response.schema";

const mockSession: AuthSessionResponse = {
  accessToken: "test-token",
  tokenType: "Bearer",
  expiresIn: 300,
  user: {
    id: "550e8400-e29b-41d4-a716-446655440000",
    email: "test@example.com",
    displayName: "Test User",
    emailVerified: true,
  },
  session: {
    id: "550e8400-e29b-41d4-a716-446655440001",
    clientId: "pdv-web",
    createdAt: "2026-07-19T12:00:00Z",
    idleExpiresAt: "2026-07-19T14:00:00Z",
    absoluteExpiresAt: "2026-07-26T12:00:00Z",
  },
  context: {
    kind: "identity",
    membershipId: null,
    organization: null,
    store: null,
    roles: [],
    scopes: [],
  },
};

describe("clearLocalAuthState", () => {
  beforeEach(() => {
    resetAccessTokenStore();
    resetAuthSessionStore();
    resetCsrfTokenStore();
    resetAuthEventBus();
  });

  afterEach(() => {
    resetAccessTokenStore();
    resetAuthSessionStore();
    resetCsrfTokenStore();
    resetAuthEventBus();
  });

  it("clears access token", () => {
    setAccessToken("my-token", 300);
    expect(getAccessToken()).toBe("my-token");
    clearLocalAuthState("logout");
    expect(getAccessToken()).toBeNull();
  });

  it("clears session state", () => {
    setAuthenticatedSession(mockSession);
    expect(getAuthSessionState().status).toBe("authenticated");
    clearLocalAuthState("logout");
    expect(getAuthSessionState().status).toBe("unknown");
    expect(getAuthSessionState().session).toBeNull();
  });

  it("clears CSRF token when set", () => {
    setCsrfToken("csrf-value");
    expect(getCsrfToken()).toBe("csrf-value");
    clearLocalAuthState("logout");
    expect(getCsrfToken()).toBeNull();
  });
});
