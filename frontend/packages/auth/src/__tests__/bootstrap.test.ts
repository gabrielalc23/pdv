import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { bootstrapAuth } from "../runtime/bootstrap";
import {
  resetAccessTokenStore,
  getAccessToken,
} from "../stores/access-token.store";
import {
  resetAuthSessionStore,
  getAuthSessionState,
} from "../stores/auth-session.store";
import { resetCsrfTokenStore } from "../stores/csrf-token.store";
import { resetAuthEventBus } from "../events/auth-events";

describe("bootstrapAuth", () => {
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

  it("returns authenticated when refresh succeeds", async () => {
    const result = await bootstrapAuth({
      publicInstance: {} as any,
      doRefresh: async () => ({
        accessToken: "new-token",
        tokenType: "Bearer" as const,
        expiresIn: 300,
        user: {
          id: "550e8400-e29b-41d4-a716-446655440000",
          email: "user@example.com",
          displayName: "User",
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
          kind: "identity" as const,
          membershipId: null,
          organization: null,
          store: null,
          roles: [],
          scopes: [],
        },
      }),
      fetchCsrf: async () => "csrf-token",
      sourceId: "test",
    });

    expect(result.status).toBe("authenticated");
    if (result.status === "authenticated") {
      expect(result.session.accessToken).toBe("new-token");
    }
    expect(getAccessToken()).toBe("new-token");
  });

  it("returns anonymous when refresh returns REFRESH_TOKEN_MISSING", async () => {
    const result = await bootstrapAuth({
      publicInstance: {} as any,
      doRefresh: async () => {
        const error = new Error("Refresh failed") as any;
        error.response = {
          status: 401,
          data: { error: { code: "REFRESH_TOKEN_MISSING" } },
        };
        throw error;
      },
      fetchCsrf: async () => "csrf-token",
      sourceId: "test",
    });

    expect(result.status).toBe("anonymous");
  });

  it("returns anonymous when session expired", async () => {
    const result = await bootstrapAuth({
      publicInstance: {} as any,
      doRefresh: async () => {
        const error = new Error("Session expired") as any;
        error.response = {
          status: 401,
          data: { error: { code: "SESSION_EXPIRED" } },
        };
        throw error;
      },
      fetchCsrf: async () => "csrf-token",
      sourceId: "test",
    });

    expect(result.status).toBe("anonymous");
  });

  it("returns unavailable on network error", async () => {
    const result = await bootstrapAuth({
      publicInstance: {} as any,
      doRefresh: async () => {
        throw new Error("Network error");
      },
      fetchCsrf: async () => "csrf-token",
      sourceId: "test",
    });

    expect(result.status).toBe("unavailable");
  });

  it("CSRF fetch failure does not prevent bootstrap", async () => {
    const result = await bootstrapAuth({
      publicInstance: {} as any,
      doRefresh: async () => ({
        accessToken: "token",
        tokenType: "Bearer" as const,
        expiresIn: 300,
        user: {
          id: "550e8400-e29b-41d4-a716-446655440000",
          email: "user@example.com",
          displayName: "User",
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
          kind: "identity" as const,
          membershipId: null,
          organization: null,
          store: null,
          roles: [],
          scopes: [],
        },
      }),
      fetchCsrf: async () => {
        throw new Error("CSRF service unavailable");
      },
      sourceId: "test",
    });

    expect(result.status).toBe("authenticated");
  });
});
