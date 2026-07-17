import { describe, it, expect } from "vitest";
import { UnauthorizedError } from "../unauthorized.error";

describe("UnauthorizedError", () => {
  it("has default message and correct status/code", () => {
    const error: UnauthorizedError = new UnauthorizedError();

    expect(error.code).toBe("unauthorized");
    expect(error.status).toBe(401);
    expect(error.message).toBe("Authentication is required");
  });

  it("accepts custom message and cause", () => {
    const cause: Error = new Error("token expired");
    const error: UnauthorizedError = new UnauthorizedError("Session expired", cause);

    expect(error.message).toBe("Session expired");
    expect(error.cause).toBe(cause);
  });
});
