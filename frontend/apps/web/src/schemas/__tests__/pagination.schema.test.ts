import { describe, it, expect } from "vitest";
import { PaginationMetaSchema } from "../pagination.schema";

describe("PaginationMetaSchema", () => {
  it("accepts valid pagination", () => {
    const result = PaginationMetaSchema.safeParse({
      page: 1,
      pageSize: 20,
      total: 100,
      totalPages: 5,
    });
    expect(result.success).toBe(true);
  });

  it("accepts zero page", () => {
    const result = PaginationMetaSchema.safeParse({
      page: 0,
      pageSize: 20,
      total: 100,
      totalPages: 5,
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing fields", () => {
    const result = PaginationMetaSchema.safeParse({ page: 1 });
    expect(result.success).toBe(false);
  });
});
