import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { logError } from "../log-error.util"

beforeEach(() => {
  vi.stubGlobal("console", { error: vi.fn() })
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("logError", () => {
  it("logs message with error prefix when no error arg", () => {
    logError("Something failed")
    expect(console.error).toHaveBeenCalledWith("❌ Something failed")
  })

  it("logs Error instance with message and stack", () => {
    const err = new Error("boom")
    logError("Failed", err)
    expect(console.error).toHaveBeenCalledWith("❌ Failed", err)
  })

  it("logs unknown value as string", () => {
    logError("Failed", 42)
    expect(console.error).toHaveBeenCalledWith("❌ Failed", "42")
  })
})
