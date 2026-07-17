import { describe, it, expect } from "vitest";
import { ValidationError } from "../validation.error";

describe("ValidationError", () => {
  it("has default message when no args given", () => {
    const error: ValidationError = new ValidationError();

    expect(error.code).toBe("validation_error");
    expect(error.status).toBe(422);
    expect(error.message).toBe("The provided data is invalid");
    expect(error.field).toBeUndefined();
    expect(error.details).toBeUndefined();
  });

  it("accepts custom message and field", () => {
    const error: ValidationError = new ValidationError({
      message: "Invalid email",
      field: "email",
    });

    expect(error.message).toBe("Invalid email");
    expect(error.field).toBe("email");
  });

  it("accepts details and cause", () => {
    const cause: Error = new Error("parse failed");
    const details: Record<string, unknown> = { minLength: 3, maxLength: 100 };
    const error: ValidationError = new ValidationError({ message: "Invalid", details, cause });

    expect(error.details).toEqual(details);
    expect(error.cause).toBe(cause);
  });
});
