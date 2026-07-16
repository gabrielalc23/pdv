import { describe, it, expect } from "vitest"
import { HealthResponseSchema } from "../health.schema"

describe("HealthResponseSchema", () => {
  it("accepts valid health response", () => {
    const result = HealthResponseSchema.safeParse({ status: "ok" })
    expect(result.success).toBe(true)
  })

  it("rejects missing status", () => {
    const result = HealthResponseSchema.safeParse({})
    expect(result.success).toBe(false)
  })
})
