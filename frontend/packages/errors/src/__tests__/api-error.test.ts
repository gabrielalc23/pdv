import { describe, it, expect } from "vitest";
import { ApiError } from "../api.error";

describe("ApiError", () => {
  it("extends AppError with correct shape", () => {
    const cause: Error = new Error("api failure");
    const error: ApiError = new ApiError("ERR_001", "API error occurred", 500, "fieldName", cause);

    expect(error).toBeInstanceOf(Error);
    expect(error.code).toBe("ERR_001");
    expect(error.message).toBe("API error occurred");
    expect(error.status).toBe(500);
    expect(error.field).toBe("fieldName");
    expect(error.cause).toBe(cause);
  });

  it("works without optional field and cause", () => {
    const error: ApiError = new ApiError("ERR_002", "Simple error", 400);

    expect(error.field).toBeUndefined();
    expect(error.cause).toBeUndefined();
  });
});
