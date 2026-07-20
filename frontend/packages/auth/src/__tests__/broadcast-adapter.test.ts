import { describe, it, expect, vi } from "vitest";
import { createBroadcastAdapter } from "../coordinator/broadcast-adapter";

describe("BroadcastAdapter", () => {
  it("creates adapter without crashing", () => {
    const adapter = createBroadcastAdapter("test-channel");
    expect(adapter).toBeDefined();
    adapter.close();
  });

  it("postMessage does not crash without channel", () => {
    const adapter = createBroadcastAdapter("test-channel");
    expect(() => adapter.postMessage({ test: true })).not.toThrow();
    adapter.close();
  });

  it("onMessage returns cleanup function", () => {
    const adapter = createBroadcastAdapter("test-channel");
    const cleanup = adapter.onMessage(() => {});
    expect(typeof cleanup).toBe("function");
    cleanup(); // should not throw
    adapter.close();
  });

  it("close is idempotent", () => {
    const adapter = createBroadcastAdapter("test-channel");
    adapter.close();
    adapter.close(); // should not throw
  });
});
