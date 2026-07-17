import { describe, it, expect } from "vitest";
import { resolveRequestLocation } from "../resolve-request-location.util";

describe("resolveRequestLocation", () => {
  it("returns explicit requestLocation when provided", () => {
    expect(resolveRequestLocation("GET", "none")).toBe("none");
    expect(resolveRequestLocation("POST", "params")).toBe("params");
  });

  it("returns 'params' for GET", () => {
    expect(resolveRequestLocation("GET")).toBe("params");
  });

  it("returns 'params' for HEAD", () => {
    expect(resolveRequestLocation("HEAD")).toBe("params");
  });

  it("returns 'none' for OPTIONS", () => {
    expect(resolveRequestLocation("OPTIONS")).toBe("none");
  });

  it("returns 'data' for POST", () => {
    expect(resolveRequestLocation("POST")).toBe("data");
  });

  it("returns 'data' for PUT", () => {
    expect(resolveRequestLocation("PUT")).toBe("data");
  });

  it("returns 'data' for DELETE", () => {
    expect(resolveRequestLocation("DELETE")).toBe("data");
  });

  it("returns 'data' for PATCH", () => {
    expect(resolveRequestLocation("PATCH")).toBe("data");
  });
});
