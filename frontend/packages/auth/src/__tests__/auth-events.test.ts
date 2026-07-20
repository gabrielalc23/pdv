import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  subscribeToAuthEvents,
  dispatchAuthEvent,
  resetAuthEventBus,
} from "../events/auth-events";

describe("AuthEventBus", () => {
  beforeEach(() => {
    resetAuthEventBus();
  });

  afterEach(() => {
    resetAuthEventBus();
  });

  it("dispatches event to subscriber", () => {
    const events: Array<string> = [];
    const unsub = subscribeToAuthEvents((event) => events.push(event.type));

    dispatchAuthEvent({ type: "token-updated" });
    expect(events).toEqual(["token-updated"]);

    dispatchAuthEvent({ type: "bootstrap-completed", authenticated: true });
    expect(events).toEqual(["token-updated", "bootstrap-completed"]);

    unsub();
  });

  it("subscriber error does not break dispatch", () => {
    const events: Array<string> = [];
    subscribeToAuthEvents(() => {
      throw new Error("bad listener");
    });
    subscribeToAuthEvents((event) => events.push(event.type));

    expect(() => dispatchAuthEvent({ type: "token-updated" })).not.toThrow();
    expect(events).toEqual(["token-updated"]);
  });

  it("unsubscribe removes listener", () => {
    const events: Array<string> = [];
    const unsub = subscribeToAuthEvents((event) => events.push(event.type));
    unsub();

    dispatchAuthEvent({ type: "token-updated" });
    expect(events).toEqual([]);
  });
});
