import { describe, it, expect } from "vitest"
import { ConflictError } from "../conflict.error"

describe("ConflictError", () => {
  it("has default message and correct status/code", () => {
    const error: ConflictError = new ConflictError()

    expect(error.code).toBe("conflict")
    expect(error.status).toBe(409)
    expect(error.message).toBe("The resource conflicts with the current state")
  })

  it("accepts custom message and cause", () => {
    const cause: Error = new Error("duplicate")
    const error: ConflictError = new ConflictError("SKU already exists", cause)

    expect(error.message).toBe("SKU already exists")
    expect(error.cause).toBe(cause)
  })
})
