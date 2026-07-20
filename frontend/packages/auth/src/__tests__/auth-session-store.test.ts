import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  getAuthSessionState,
  setAuthenticatedSession,
  setAnonymousSession,
  resetAuthSession,
  subscribeToAuthSession,
  resetAuthSessionStore,
} from "../stores/auth-session.store";
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

describe("AuthSessionStore", () => {
  beforeEach(() => {
    resetAuthSessionStore();
  });

  afterEach(() => {
    resetAuthSessionStore();
  });

  it("starts with unknown status", () => {
    const state = getAuthSessionState();
    expect(state.status).toBe("unknown");
    expect(state.session).toBeNull();
  });

  it("setAuthenticatedSession updates status and session", () => {
    setAuthenticatedSession(mockSession);
    const state = getAuthSessionState();
    expect(state.status).toBe("authenticated");
    expect(state.session).toEqual(mockSession);
  });

  it("setAnonymousSession resets to anonymous", () => {
    setAuthenticatedSession(mockSession);
    setAnonymousSession();
    const state = getAuthSessionState();
    expect(state.status).toBe("anonymous");
    expect(state.session).toBeNull();
  });

  it("resetAuthSession returns to unknown", () => {
    setAuthenticatedSession(mockSession);
    resetAuthSession();
    const state = getAuthSessionState();
    expect(state.status).toBe("unknown");
    expect(state.session).toBeNull();
  });

  it("notifies subscribers", () => {
    const updates: Array<string> = [];
    const unsub = subscribeToAuthSession((state) => updates.push(state.status));

    setAuthenticatedSession(mockSession);
    expect(updates).toEqual(["authenticated"]);

    setAnonymousSession();
    expect(updates).toEqual(["authenticated", "anonymous"]);

    unsub();
  });

  it("subscriber error does not corrupt state", () => {
    const unsub = subscribeToAuthSession(() => {
      throw new Error("subscriber error");
    });

    expect(() => setAuthenticatedSession(mockSession)).not.toThrow();
    expect(getAuthSessionState().status).toBe("authenticated");

    unsub();
  });
});
