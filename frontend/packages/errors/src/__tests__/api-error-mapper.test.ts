import { describe, it, expect, vi } from "vitest";
import { mapApiError } from "../api-error-mapper";
import { AppError } from "../app.error";
import { ApiError } from "../api.error";
import { NotFoundError } from "../not-found.error";
import { ConflictError } from "../conflict.error";
import { ValidationError } from "../validation.error";

interface MockAxiosResponse {
  status: number;
  data: unknown;
}

type MockAxiosError = Error & {
  isAxiosError: true;
  response: MockAxiosResponse;
};

function createAxiosError(status: number, body?: unknown): MockAxiosError {
  return Object.assign(new Error("Axios error"), {
    isAxiosError: true as const,
    response: {
      status,
      data: body,
    },
  });
}
describe("mapApiError", () => {
  it("re-throws AppError instances", () => {
    const original: AppError = new AppError({ code: "X", message: "X", status: 400 });
    expect(() => mapApiError(original)).toThrow(original);
  });

  it("throws original error for non-Axios errors", () => {
    const original = new Error("network");
    expect(() => mapApiError(original)).toThrow(original);
  });

  it("throws original error for Axios errors without error body", () => {
    const axiosErr = createAxiosError(500);
    expect(() => mapApiError(axiosErr)).toThrow(axiosErr);
  });

  it("maps 400 to ValidationError", () => {
    const axiosErr = createAxiosError(400, {
      error: { code: "invalid", message: "Bad request", field: "email" },
    });
    expect(() => mapApiError(axiosErr)).toThrow(ValidationError);
    try {
      mapApiError(axiosErr);
    } catch (e: any) {
      expect(e.field).toBe("email");
    }
  });

  it("maps 422 to ValidationError", () => {
    const axiosErr = createAxiosError(422, {
      error: { code: "validation", message: "Invalid data", field: "name" },
    });
    expect(() => mapApiError(axiosErr)).toThrow(ValidationError);
    try {
      mapApiError(axiosErr);
    } catch (e: any) {
      expect(e.message).toBe("Invalid data");
    }
  });

  it("maps 404 to NotFoundError", () => {
    const axiosErr = createAxiosError(404, {
      error: { code: "not_found", message: "Product not found" },
    });
    expect(() => mapApiError(axiosErr)).toThrow(NotFoundError);
  });

  it("maps 409 to ConflictError", () => {
    const axiosErr = createAxiosError(409, { error: { code: "conflict", message: "SKU exists" } });
    expect(() => mapApiError(axiosErr)).toThrow(ConflictError);
  });

  it("maps unknown status to generic ApiError", () => {
    const axiosErr = createAxiosError(500, {
      error: { code: "server_error", message: "Internal error" },
    });
    expect(() => mapApiError(axiosErr)).toThrow(ApiError);
    try {
      mapApiError(axiosErr);
    } catch (e: any) {
      expect(e.code).toBe("server_error");
      expect(e.status).toBe(500);
    }
  });

  it("passes cause from original Axios error", () => {
    const axiosErr = createAxiosError(404, { error: { code: "not_found", message: "Not found" } });
    expect(() => mapApiError(axiosErr)).toThrow(NotFoundError);
    try {
      mapApiError(axiosErr);
    } catch (e: any) {
      expect(e.cause).toBe(axiosErr);
    }
  });

  it("defaults status to 500 when response is missing", () => {
    const axiosErr = new Error("no response") as any;
    axiosErr.isAxiosError = true;
    expect(() => mapApiError(axiosErr)).toThrow(axiosErr);
  });
});
