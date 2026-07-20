import { describe, it, expect, beforeEach, afterEach } from "vitest";
import type { Clock } from "../stores/access-token.store";
import {
  getAccessToken,
  getAccessTokenState,
  setAccessToken,
  clearAccessToken,
  subscribeToAccessToken,
  resetAccessTokenStore,
  setClock,
} from "../stores/access-token.store";

describe("AccessTokenStore", () => {
  beforeEach(() => {
    resetAccessTokenStore();
  });

  afterEach(() => {
    resetAccessTokenStore();
  });

  it("starts with null token and expiresAt", () => {
    expect(getAccessToken()).toBeNull();
    const state = getAccessTokenState();
    expect(state.token).toBeNull();
    expect(state.expiresAt).toBeNull();
  });

  it("sets token with expiresAt computed from clock", () => {
    setAccessToken("test-token", 300);
    expect(getAccessToken()).toBe("test-token");
    const state = getAccessTokenState();
    expect(state.token).toBe("test-token");
    expect(state.expiresAt).toBeGreaterThan(Date.now());
  });

  it("rejects empty token", () => {
    expect(() => setAccessToken("", 300)).toThrow();
  });

  it("rejects zero expiresIn", () => {
    expect(() => setAccessToken("token", 0)).toThrow();
  });

  it("rejects negative expiresIn", () => {
    expect(() => setAccessToken("token", -1)).toThrow();
  });

  it("clear resets token and expiresAt", () => {
    setAccessToken("test-token", 300);
    clearAccessToken();
    expect(getAccessToken()).toBeNull();
    const state = getAccessTokenState();
    expect(state.token).toBeNull();
    expect(state.expiresAt).toBeNull();
  });

  it("subscribes and receives updates", () => {
    const updates: Array<string | null> = [];
    const unsubscribe = subscribeToAccessToken((state) => {
      updates.push(state.token);
    });

    setAccessToken("token-1", 300);
    expect(updates).toEqual(["token-1"]);

    setAccessToken("token-2", 300);
    expect(updates).toEqual(["token-1", "token-2"]);

    unsubscribe();
  });

  it("unsubscribe is idempotent", () => {
    const updates: Array<string | null> = [];
    const unsubscribe = subscribeToAccessToken((state) => {
      updates.push(state.token);
    });

    setAccessToken("token-1", 300);
    unsubscribe();
    unsubscribe(); // should not throw

    setAccessToken("token-2", 300);
    expect(updates).toEqual(["token-1"]);
  });

  it("supports multiple subscribers", () => {
    const updates1: Array<string | null> = [];
    const updates2: Array<string | null> = [];

    const unsub1 = subscribeToAccessToken((s) => updates1.push(s.token));
    const unsub2 = subscribeToAccessToken((s) => updates2.push(s.token));

    setAccessToken("multi-token", 300);
    expect(updates1).toEqual(["multi-token"]);
    expect(updates2).toEqual(["multi-token"]);

    unsub1();
    unsub2();
  });

  it("snapshot is immutable", () => {
    setAccessToken("snapshot-test", 300);
    const state1 = getAccessTokenState();
    const state2 = getAccessTokenState();
    expect(state1).not.toBe(state2); // different frozen objects
  });

  it("uses injected clock", () => {
    const fixedTime = 1_000_000;
    const mockClock: Clock = { now: () => fixedTime };
    setClock(mockClock);

    setAccessToken("timed-token", 300);
    const state = getAccessTokenState();
    expect(state.expiresAt).toBe(fixedTime + 300_000);
  });

  it("clears between resets", () => {
    setAccessToken("first", 300);
    resetAccessTokenStore();
    expect(getAccessToken()).toBeNull();
    expect(getAccessTokenState().expiresAt).toBeNull();

    setAccessToken("second", 300);
    expect(getAccessToken()).toBe("second");
  });
});
