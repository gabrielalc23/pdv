import { describe, it, expect } from "vitest"
import {
  CreateInventoryEntryInputSchema,
  CreateInventoryAdjustmentInputSchema,
  InventoryResponseSchema,
} from "../inventory.schema"

describe("CreateInventoryEntryInputSchema", () => {
  it("accepts valid entry", () => {
    const result = CreateInventoryEntryInputSchema.safeParse({
      productId: "550e8400-e29b-41d4-a716-446655440000",
      quantity: "10.000",
      reason: null,
      referenceType: "PURCHASE",
      referenceId: "550e8400-e29b-41d4-a716-446655440001",
    })
    expect(result.success).toBe(true)
  })

  it("accepts entry with reason", () => {
    const result = CreateInventoryEntryInputSchema.safeParse({
      productId: "uuid",
      quantity: "5.000",
      reason: "Compra de reposição",
      referenceType: "PURCHASE",
      referenceId: "uuid",
    })
    expect(result.success).toBe(true)
  })

  it("rejects missing productId", () => {
    const result = CreateInventoryEntryInputSchema.safeParse({
      quantity: "10.000",
      referenceType: "PURCHASE",
      referenceId: "uuid",
    })
    expect(result.success).toBe(false)
  })
})

describe("CreateInventoryAdjustmentInputSchema", () => {
  it("accepts valid adjustment IN", () => {
    const result = CreateInventoryAdjustmentInputSchema.safeParse({
      productId: "uuid",
      direction: "IN",
      quantity: "5.000",
      reason: "Ajuste manual",
      referenceType: "MANUAL",
      referenceId: "uuid",
    })
    expect(result.success).toBe(true)
  })

  it("accepts valid adjustment OUT", () => {
    const result = CreateInventoryAdjustmentInputSchema.safeParse({
      productId: "uuid",
      direction: "OUT",
      quantity: "3.000",
      reason: "Quebra",
      referenceType: "LOSS",
      referenceId: "uuid",
    })
    expect(result.success).toBe(true)
  })

  it("rejects invalid direction", () => {
    const result = CreateInventoryAdjustmentInputSchema.safeParse({
      productId: "uuid",
      direction: "UP",
      quantity: "1.000",
      reason: "Teste",
      referenceType: "TEST",
      referenceId: "uuid",
    })
    expect(result.success).toBe(false)
  })
})

describe("InventoryResponseSchema", () => {
  it("accepts response without barcode", () => {
    const result = InventoryResponseSchema.safeParse({
      productId: "uuid",
      sku: "ABC",
      name: "Teste",
      quantity: "50.000",
      isActive: true,
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(true)
  })

  it("accepts response with barcode", () => {
    const result = InventoryResponseSchema.safeParse({
      productId: "uuid",
      sku: "ABC",
      barcode: "7891234567890",
      name: "Teste",
      quantity: "50.000",
      isActive: true,
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(true)
  })
})
