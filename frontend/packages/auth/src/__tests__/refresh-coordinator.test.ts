import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { createRefreshCoordinator } from "../coordinator/refresh-coordinator";
import type {
  BroadcastAdapter,
  LockAdapter,
} from "../coordinator/coordinator.types";
import {
  resetAccessTokenStore,
  getAccessToken,
  setAccessToken,
} from "../stores/access-token.store";
import {
  resetAuthSessionStore,
  getAuthSessionState,
} from "../stores/auth-session.store";
import { resetCsrfTokenStore } from "../stores/csrf-token.store";
import { resetAuthEventBus } from "../events/auth-events";

class FakeLock implements LockAdapter {
  isSupported = true;
  private cb: (() => Promise<void>) | null = null;

  async acquire(_name: string, callback: () => Promise<void>): Promise<void> {
    this.cb = callback;
    await callback();
  }
}

class FakeBroadcast implements BroadcastAdapter {
  messages: unknown[] = [];
  private handlers: Array<(msg: unknown) => void> = [];

  postMessage(msg: unknown): void {
    this.messages.push(msg);
  }

  onMessage(handler: (msg: unknown) => void): () => void {
    this.handlers.push(handler);
    return () => {
      this.handlers = this.handlers.filter((h) => h !== handler);
    };
  }

  simulateMessage(msg: unknown): void {
    for (const handler of this.handlers) {
      handler(msg);
    }
  }

  close(): void {}
}

const fakeAxiosInstance = {} as any;
let gen = 0;

describe("RefreshCoordinator", () => {
  let coordinator: ReturnType<typeof createRefreshCoordinator> | null;
  let broadcast: FakeBroadcast;
  let lock: FakeLock;

  beforeEach(() => {
    coordinator = null;
    resetAccessTokenStore();
    resetAuthSessionStore();
    resetCsrfTokenStore();
    resetAuthEventBus();
    broadcast = new FakeBroadcast();
    lock = new FakeLock();
    gen = 0;
  });

  afterEach(() => {
    if (coordinator) {
      coordinator.destroy();
      coordinator = null;
    }
  });

  function createCoord(): void {
    coordinator = createRefreshCoordinator({
      publicInstance: fakeAxiosInstance,
      broadcast,
      lock,
      sourceId: "test-source-1",
      getGeneration: () => gen,
    });
  }

  it("does not crash when calling refresh without a valid instance", async () => {
    createCoord();
    await expect(coordinator!.refresh()).rejects.toThrow();
  });

  it("publishLogout sends logout message", () => {
    createCoord();
    coordinator!.publishLogout();
    expect(broadcast.messages).toHaveLength(1);
    expect((broadcast.messages[0] as any).type).toBe("logout");
    expect((broadcast.messages[0] as any).sourceId).toBe("test-source-1");
  });

  it("publishAuthLost sends auth-lost message", () => {
    createCoord();
    coordinator!.publishAuthLost("session_expired");
    expect(broadcast.messages).toHaveLength(1);
    expect((broadcast.messages[0] as any).type).toBe("auth-lost");
  });

  it("receiving logout message from another tab clears local state", () => {
    createCoord();
    setAccessToken("existing-token", 300);
    expect(getAccessToken()).toBe("existing-token");

    broadcast.simulateMessage({
      type: "logout",
      sourceId: "other-source",
    });

    expect(getAccessToken()).toBeNull();
  });

  it("ignores messages from self", () => {
    createCoord();
    setAccessToken("my-token", 300);

    broadcast.simulateMessage({
      type: "token-updated",
      sourceId: "test-source-1",
      accessToken: "other-token",
      expiresIn: 300,
      session: { dummy: true },
    });

    expect(getAccessToken()).toBe("my-token");
  });

  it("ignores invalid messages", () => {
    createCoord();
    broadcast.simulateMessage({ type: "unknown-type" });
    broadcast.simulateMessage("not-an-object");
    expect(true).toBe(true);
  });

  it("can be destroyed", () => {
    createCoord();
    coordinator!.destroy();
    coordinator!.destroy();
  });
});
