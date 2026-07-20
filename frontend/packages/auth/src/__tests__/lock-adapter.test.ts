import { describe, it, expect } from "vitest";
import { createLockAdapter } from "../coordinator/lock-adapter";

describe("LockAdapter", () => {
  it("creates adapter without crashing", () => {
    const adapter = createLockAdapter();
    expect(adapter).toBeDefined();
    expect(typeof adapter.isSupported).toBe("boolean");
  });

  it("acquire works without navigator.locks", async () => {
    const adapter = createLockAdapter();
    let hasCalled = false;
    await adapter.acquire("test-lock", async () => {
      hasCalled = true;
    });
    expect(hasCalled).toBe(true);
  });
});
