import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { log } from "../log.util"

beforeEach(() => {
  vi.stubGlobal("console", { log: vi.fn() })
  vi.stubGlobal("navigator", { sendBeacon: vi.fn() })
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("log", () => {
  it("logs to console in development", () => {
    vi.stubEnv("NODE_ENV", "development")
    log("hello")
    expect(console.log).toHaveBeenCalledWith("hello")
  })

  it("sends beacon in production", () => {
    vi.stubEnv("NODE_ENV", "production")
    log("hello")
    expect(navigator.sendBeacon).toHaveBeenCalledWith("/api/log", "hello")
    expect(console.log).not.toHaveBeenCalled()
  })

  it("does not call console.log in production", () => {
    vi.stubEnv("NODE_ENV", "production")
    log("prod msg")
    expect(console.log).not.toHaveBeenCalled()
  })

  it("does not throw if navigator is undefined in production", () => {
    vi.stubEnv("NODE_ENV", "production")
    vi.stubGlobal("navigator", undefined)
    expect(() => log("safe")).not.toThrow()
  })
})
