import { describe, it, expect } from "vitest";
import { AppError } from "../app.error";

describe("AppError", () => {
  it("creates error with all fields", () => {
    const cause: Error = new Error("root");
    const error: AppError = new AppError({
      code: "test_code",
      message: "Test message",
      status: 400,
      field: "name",
      cause,
    });

    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe("AppError");
    expect(error.code).toBe("test_code");
    expect(error.message).toBe("Test message");
    expect(error.status).toBe(400);
    expect(error.field).toBe("name");
    expect(error.cause).toBe(cause);
  });

  it("creates error without optional fields", () => {
    const error: AppError = new AppError({ code: "test", message: "msg", status: 500 });

    expect(error.field).toBeUndefined();
    expect(error.cause).toBeUndefined();
  });
});
