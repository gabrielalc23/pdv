import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  getCsrfToken,
  setCsrfToken,
  clearCsrfToken,
  subscribeToCsrfToken,
  resetCsrfTokenStore,
} from "../stores/csrf-token.store";

describe("CsrfTokenStore", () => {
  beforeEach(() => {
    resetCsrfTokenStore();
  });

  afterEach(() => {
    resetCsrfTokenStore();
  });

  it("starts with null", () => {
    expect(getCsrfToken()).toBeNull();
  });

  it("sets and retrieves token", () => {
    setCsrfToken("csrf-token-value");
    expect(getCsrfToken()).toBe("csrf-token-value");
  });

  it("clears token", () => {
    setCsrfToken("some-token");
    clearCsrfToken();
    expect(getCsrfToken()).toBeNull();
  });

  it("notifies subscribers on set", () => {
    const updates: Array<string | null> = [];
    const unsub = subscribeToCsrfToken((token) => updates.push(token));

    setCsrfToken("new-token");
    expect(updates).toEqual(["new-token"]);

    unsub();
  });

  it("notifies subscribers on clear", () => {
    setCsrfToken("before");
    const updates: Array<string | null> = [];
    const unsub = subscribeToCsrfToken((token) => updates.push(token));

    clearCsrfToken();
    expect(updates).toEqual([null]);

    unsub();
  });

  it("unsubscribe is idempotent", () => {
    const updates: Array<string | null> = [];
    const unsub = subscribeToCsrfToken((token) => updates.push(token));
    unsub();
    unsub();

    setCsrfToken("should-not-appear");
    expect(updates).toEqual([]);
  });
});
