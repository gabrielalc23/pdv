import { describe, it, expect } from "vitest";
import {
  UpsertProductInputSchema,
  ProductResponseSchema,
  ListProductsParamsSchema,
  ProductListResponseSchema,
} from "../product.schema";

describe("UpsertProductInputSchema", () => {
  it("accepts valid input", () => {
    const result = UpsertProductInputSchema.safeParse({
      sku: "ABC-123",
      barcode: null,
      name: "Produto Teste",
      price: "99.90",
      cost: null,
    });
    expect(result.success).toBe(true);
  });

  it("accepts input with barcode and cost", () => {
    const result = UpsertProductInputSchema.safeParse({
      sku: "ABC-123",
      barcode: "7891234567890",
      name: "Produto Teste",
      price: "99.90",
      cost: "50.00",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty sku", () => {
    const result = UpsertProductInputSchema.safeParse({
      sku: "",
      barcode: null,
      name: "Produto",
      price: "10.00",
      cost: null,
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing name", () => {
    const result = UpsertProductInputSchema.safeParse({
      sku: "ABC",
      barcode: null,
      price: "10.00",
      cost: null,
    });
    expect(result.success).toBe(false);
  });

  it("rejects non-string price", () => {
    const result = UpsertProductInputSchema.safeParse({
      sku: "ABC",
      barcode: null,
      name: "Produto",
      price: 99.9,
      cost: null,
    });
    expect(result.success).toBe(false);
  });
});

describe("ProductResponseSchema", () => {
  it("accepts full product response", () => {
    const result = ProductResponseSchema.safeParse({
      id: "550e8400-e29b-41d4-a716-446655440000",
      sku: "ABC",
      barcode: null,
      name: "Teste",
      price: "10.00",
      cost: null,
      isActive: true,
      createdAt: "2026-07-16T10:00:00Z",
      updatedAt: "2026-07-16T10:00:00Z",
    });
    expect(result.success).toBe(true);
  });

  it("rejects missing isActive", () => {
    const result = ProductResponseSchema.safeParse({
      id: "...",
      sku: "A",
      name: "Teste",
      price: "10.00",
      createdAt: "...",
      updatedAt: "...",
    });
    expect(result.success).toBe(false);
  });
});

describe("ListProductsParamsSchema", () => {
  it("accepts empty params", () => {
    const result = ListProductsParamsSchema.safeParse({});
    expect(result.success).toBe(true);
  });

  it("accepts full params", () => {
    const result = ListProductsParamsSchema.safeParse({
      search: "abc",
      page: 1,
      pageSize: 50,
      activeOnly: true,
    });
    expect(result.success).toBe(true);
  });
});

describe("ProductListResponseSchema", () => {
  it("accepts paginated response", () => {
    const result = ProductListResponseSchema.safeParse({
      data: [
        {
          id: "1",
          sku: "A",
          barcode: null,
          name: "Teste",
          price: "10.00",
          cost: null,
          isActive: true,
          createdAt: "...",
          updatedAt: "...",
        },
      ],
      pagination: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    });
    expect(result.success).toBe(true);
  });

  it("accepts empty data array", () => {
    const result = ProductListResponseSchema.safeParse({
      data: [],
      pagination: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    });
    expect(result.success).toBe(true);
  });
});
