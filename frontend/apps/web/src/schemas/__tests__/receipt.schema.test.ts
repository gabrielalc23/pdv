import { describe, it, expect } from "vitest"
import { ReceiptResponseSchema } from "../receipt.schema"

describe("ReceiptResponseSchema", () => {
  it("accepts complete receipt", () => {
    const result = ReceiptResponseSchema.safeParse({
      sale: {
        id: "uuid",
        number: 1,
        status: "COMPLETED",
        subtotal: "100.00",
        discount: "0.00",
        addition: "0.00",
        total: "100.00",
        openedAt: "2026-07-16T10:00:00Z",
        completedAt: "2026-07-16T10:05:00Z",
        cancelledAt: null,
        createdAt: "2026-07-16T10:00:00Z",
        updatedAt: "2026-07-16T10:05:00Z",
        idempotencyKey: "key-123",
      },
      items: [
        {
          productId: "uuid",
          sku: "ABC",
          name: "Produto",
          unitPrice: "10.00",
          quantity: "2.000",
          subtotal: "20.00",
          discount: "0.00",
          total: "20.00",
          createdAt: "...",
        },
      ],
      payments: [
        {
          method: "Pix",
          amount: "100.00",
          status: "APPROVED",
          installments: 1,
        },
      ],
      fiscalDocument: {
        status: "AUTHORIZED",
        accessKey: "35200600000000000000550000000000000000000000",
        protocol: "MOCK-123",
        provider: "mock",
        externalReference: "sale-uuid",
      },
    })
    expect(result.success).toBe(true)
  })

  it("accepts receipt without fiscal document", () => {
    const result = ReceiptResponseSchema.safeParse({
      sale: {
        id: "uuid",
        number: 1,
        status: "COMPLETED",
        subtotal: "50.00",
        discount: "0.00",
        addition: "0.00",
        total: "50.00",
        openedAt: "...",
        completedAt: "...",
        cancelledAt: null,
        createdAt: "...",
        updatedAt: "...",
        idempotencyKey: "key",
      },
      items: [],
      payments: [],
      fiscalDocument: null,
    })
    expect(result.success).toBe(true)
  })
})
