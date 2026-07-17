import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { sendLog } from "../send-log.util";

beforeEach(() => {
  vi.stubGlobal("navigator", { sendBeacon: vi.fn() });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("sendLog", () => {
  it("calls sendBeacon with serialized JSON blob", () => {
    const payload = { event: "click", page: "/home" };
    sendLog(payload);

    expect(navigator.sendBeacon).toHaveBeenCalledTimes(1);
    const [url, blob] = (navigator.sendBeacon as any).mock.calls[0];
    expect(url).toBe("/api/log-endpoint");
    expect(blob).toBeInstanceOf(Blob);
    expect(blob.type).toBe("application/json");
  });
});
