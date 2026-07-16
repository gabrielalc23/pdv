import { describe, it, expect } from "vitest"
import { InvalidApiResponseError } from "../invalid-api-response.error"

describe("InvalidApiResponseError", () => {
  it("includes path in message and has correct status/code", () => {
    const error: InvalidApiResponseError = new InvalidApiResponseError("/products")

    expect(error.code).toBe("invalid_api_response")
    expect(error.status).toBe(502)
    expect(error.message).toContain("/products")
    expect(error.message).toContain("does not match the expected schema")
  })

  it("accepts cause", () => {
    const cause: Error = new Error("zod parse failed")
    const error: InvalidApiResponseError = new InvalidApiResponseError("/sales", cause)

    expect(error.cause).toBe(cause)
  })
})
