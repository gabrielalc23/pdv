import { describe, it, expect } from "vitest"
import { ListCatalogParamsSchema, CatalogProductResponseSchema } from "../catalog.schema"

describe("ListCatalogParamsSchema", () => {
  it("accepts empty params", () => {
    const result = ListCatalogParamsSchema.safeParse({})
    expect(result.success).toBe(true)
  })

  it("accepts inStockOnly filter", () => {
    const result = ListCatalogParamsSchema.safeParse({ inStockOnly: true })
    expect(result.success).toBe(true)
  })
})

describe("CatalogProductResponseSchema", () => {
  it("accepts valid catalog product", () => {
    const result = CatalogProductResponseSchema.safeParse({
      id: "uuid",
      sku: "ABC",
      barcode: null,
      name: "Produto",
      price: "29.90",
      quantity: "100.000",
      isActive: true,
      inStock: true,
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing inStock", () => {
    const result = CatalogProductResponseSchema.safeParse({
      id: "uuid",
      sku: "ABC",
      name: "Produto",
      price: "10.00",
      quantity: "0.000",
      isActive: true,
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(false)
  })
})
