import { describe, it, expect } from "vitest";
import { ForbiddenError } from "../forbidden.error";

describe("ForbiddenError", () => {
  it("has default message and correct status/code", () => {
    const error: ForbiddenError = new ForbiddenError();

    expect(error.code).toBe("forbidden");
    expect(error.status).toBe(403);
    expect(error.message).toBe("You do not have permission to perform this action");
  });

  it("accepts custom message and cause", () => {
    const cause: Error = new Error("no access");
    const error: ForbiddenError = new ForbiddenError("Admin only", cause);

    expect(error.message).toBe("Admin only");
    expect(error.cause).toBe(cause);
  });
});
