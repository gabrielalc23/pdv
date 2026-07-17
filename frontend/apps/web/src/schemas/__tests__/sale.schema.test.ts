import { describe, it, expect } from "vitest";
import {
  CreateSaleInputSchema,
  ListSalesParamsSchema,
  AddSaleItemInputSchema,
  UpdateSaleItemInputSchema,
  SaleResponseSchema,
  SaleListResponseSchema,
} from "../sale.schema";

describe("CreateSaleInputSchema", () => {
  it("accepts valid input", () => {
    const result = CreateSaleInputSchema.safeParse({
      idempotencyKey: "key-123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing idempotencyKey", () => {
    const result = CreateSaleInputSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});

describe("ListSalesParamsSchema", () => {
  it("accepts empty params", () => {
    const result = ListSalesParamsSchema.safeParse({});
    expect(result.success).toBe(true);
  });

  it("accepts valid status filter", () => {
    const result = ListSalesParamsSchema.safeParse({ status: "OPEN" });
    expect(result.success).toBe(true);
  });

  it("rejects invalid status", () => {
    const result = ListSalesParamsSchema.safeParse({ status: "INVALID" });
    expect(result.success).toBe(false);
  });
});

describe("AddSaleItemInputSchema", () => {
  it("accepts valid input without discount", () => {
    const result = AddSaleItemInputSchema.safeParse({
      productId: "uuid",
      quantity: "2.000",
    });
    expect(result.success).toBe(true);
  });

  it("accepts input with discount", () => {
    const result = AddSaleItemInputSchema.safeParse({
      productId: "uuid",
      quantity: "1.000",
      discount: "5.00",
    });
    expect(result.success).toBe(true);
  });
});

describe("UpdateSaleItemInputSchema", () => {
  it("accepts valid update", () => {
    const result = UpdateSaleItemInputSchema.safeParse({
      quantity: "3.000",
      discount: null,
    });
    expect(result.success).toBe(true);
  });
});

describe("SaleResponseSchema", () => {
  it("accepts open sale with items", () => {
    const result = SaleResponseSchema.safeParse({
      id: "uuid",
      number: 1,
      status: "OPEN",
      subtotal: "100.00",
      discount: "0.00",
      addition: "0.00",
      total: "100.00",
      openedAt: "2026-07-16T10:00:00Z",
      completedAt: null,
      cancelledAt: null,
      createdAt: "2026-07-16T10:00:00Z",
      updatedAt: "2026-07-16T10:00:00Z",
      idempotencyKey: "key-123",
      items: [],
    });
    expect(result.success).toBe(true);
  });
});

describe("SaleListResponseSchema", () => {
  it("accepts paginated response", () => {
    const result = SaleListResponseSchema.safeParse({
      data: [],
      pagination: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    });
    expect(result.success).toBe(true);
  });
});
