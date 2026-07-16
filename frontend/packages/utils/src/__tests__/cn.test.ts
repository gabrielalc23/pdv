import { describe, it, expect } from "vitest"
import { cn } from "../cn.util"

describe("cn", () => {
  it("merges classes with conditional values", () => {
    const result = cn("px-4", "py-2", false && "hidden", true && "flex")
    expect(result).toBe("px-4 py-2 flex")
  })

  it("resolves Tailwind conflicts (last wins)", () => {
    const result = cn("px-4", "px-6")
    expect(result).toBe("px-6")
  })

  it("handles empty inputs", () => {
    expect(cn()).toBe("")
  })

  it("filters out undefined and null, keeps non-conflicting classes", () => {
    const result = cn("px-4", undefined, null, "py-2")
    expect(result).not.toContain("undefined")
    expect(result).not.toContain("null")
    expect(result).toContain("px-4")
    expect(result).toContain("py-2")
  })

  it("handles object syntax", () => {
    expect(cn({ "bg-red-500": true, "bg-blue-500": false })).toBe("bg-red-500")
  })
})
