import { describe, it, expect } from "vitest";
import { NotFoundError } from "../not-found.error";

describe("NotFoundError", () => {
  it("has default message and correct status/code", () => {
    const error: NotFoundError = new NotFoundError();

    expect(error.code).toBe("not_found");
    expect(error.status).toBe(404);
    expect(error.message).toBe("Resource not found");
  });

  it("accepts custom message and cause", () => {
    const cause: Error = new Error("db miss");
    const error: NotFoundError = new NotFoundError("Product not found", cause);

    expect(error.message).toBe("Product not found");
    expect(error.cause).toBe(cause);
  });
});
