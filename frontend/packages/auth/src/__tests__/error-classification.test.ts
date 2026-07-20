import { describe, it, expect } from "vitest";
import {
  isRefreshEligibleError,
  isTerminalAuthError,
  isExpectedAnonymousRefreshError,
} from "../runtime/configure-http";
import type { ApiErrorLike } from "../types/transport.types";

describe("isRefreshEligibleError", () => {
  it("returns true for ACCESS_TOKEN_EXPIRED", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "ACCESS_TOKEN_EXPIRED" }),
    ).toBe(true);
  });

  it("returns true for AUTHORIZATION_STALE", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "AUTHORIZATION_STALE" }),
    ).toBe(true);
  });

  it("returns true for AUTH_CONTEXT_STALE", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "AUTH_CONTEXT_STALE" }),
    ).toBe(true);
  });

  it("returns false for INVALID_CREDENTIALS", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "INVALID_CREDENTIALS" }),
    ).toBe(false);
  });

  it("returns false for ACCESS_TOKEN_MISSING", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "ACCESS_TOKEN_MISSING" }),
    ).toBe(false);
  });

  it("returns false for ACCESS_TOKEN_INVALID", () => {
    expect(
      isRefreshEligibleError({ status: 401, code: "ACCESS_TOKEN_INVALID" }),
    ).toBe(false);
  });

  it("returns false for INSUFFICIENT_SCOPE", () => {
    expect(
      isRefreshEligibleError({ status: 403, code: "INSUFFICIENT_SCOPE" }),
    ).toBe(false);
  });

  it("returns false for VALIDATION_ERROR", () => {
    expect(
      isRefreshEligibleError({ status: 422, code: "VALIDATION_ERROR" }),
    ).toBe(false);
  });

  it("returns false for CSRF_INVALID", () => {
    expect(isRefreshEligibleError({ status: 403, code: "CSRF_INVALID" })).toBe(
      false,
    );
  });

  it("returns false for RATE_LIMITED", () => {
    expect(isRefreshEligibleError({ status: 429, code: "RATE_LIMITED" })).toBe(
      false,
    );
  });

  it("returns false for 403 generic", () => {
    expect(isRefreshEligibleError({ status: 403, code: "FORBIDDEN" })).toBe(
      false,
    );
  });

  it("returns false for 404 generic", () => {
    expect(isRefreshEligibleError({ status: 404, code: "NOT_FOUND" })).toBe(
      false,
    );
  });

  it("returns false when code is null", () => {
    expect(isRefreshEligibleError({ status: 401, code: null })).toBe(false);
  });
});

describe("isTerminalAuthError", () => {
  it("returns true for SESSION_REVOKED", () => {
    expect(isTerminalAuthError({ status: 401, code: "SESSION_REVOKED" })).toBe(
      true,
    );
  });

  it("returns true for SESSION_EXPIRED", () => {
    expect(isTerminalAuthError({ status: 401, code: "SESSION_EXPIRED" })).toBe(
      true,
    );
  });

  it("returns true for REFRESH_TOKEN_REUSED", () => {
    expect(
      isTerminalAuthError({ status: 401, code: "REFRESH_TOKEN_REUSED" }),
    ).toBe(true);
  });

  it("returns true for REFRESH_TOKEN_INVALID", () => {
    expect(
      isTerminalAuthError({ status: 401, code: "REFRESH_TOKEN_INVALID" }),
    ).toBe(true);
  });

  it("returns true for REFRESH_TOKEN_EXPIRED", () => {
    expect(
      isTerminalAuthError({ status: 401, code: "REFRESH_TOKEN_EXPIRED" }),
    ).toBe(true);
  });

  it("returns false for ACCESS_TOKEN_EXPIRED", () => {
    expect(
      isTerminalAuthError({ status: 401, code: "ACCESS_TOKEN_EXPIRED" }),
    ).toBe(false);
  });

  it("returns false for non-terminal codes", () => {
    expect(
      isTerminalAuthError({ status: 401, code: "INVALID_CREDENTIALS" }),
    ).toBe(false);
  });

  it("returns false when code is null", () => {
    expect(isTerminalAuthError({ status: 401, code: null })).toBe(false);
  });
});

describe("isExpectedAnonymousRefreshError", () => {
  it("returns true for REFRESH_TOKEN_MISSING", () => {
    expect(
      isExpectedAnonymousRefreshError({
        status: 401,
        code: "REFRESH_TOKEN_MISSING",
      }),
    ).toBe(true);
  });

  it("returns false for other codes", () => {
    expect(
      isExpectedAnonymousRefreshError({ status: 401, code: "SESSION_EXPIRED" }),
    ).toBe(false);
  });
});
