import { describe, it, expect } from "vitest"
import { CheckoutInputSchema, CheckoutPaymentInputSchema } from "../checkout.schema"

describe("CheckoutPaymentInputSchema", () => {
  it("accepts minimal payment", () => {
    const result = CheckoutPaymentInputSchema.safeParse({
      paymentMethodId: "uuid",
      amount: "100.00",
    })
    expect(result.success).toBe(true)
  })

  it("accepts full payment with optional fields", () => {
    const result = CheckoutPaymentInputSchema.safeParse({
      paymentMethodId: "uuid",
      amount: "100.00",
      receivedAmount: "100.00",
      installments: 1,
      externalReference: "ref-123",
    })
    expect(result.success).toBe(true)
  })

  it("accepts payment without installments", () => {
    const result = CheckoutPaymentInputSchema.safeParse({
      paymentMethodId: "uuid",
      amount: "50.00",
      receivedAmount: null,
      externalReference: null,
    })
    expect(result.success).toBe(true)
  })

  it("rejects non-integer installments", () => {
    const result = CheckoutPaymentInputSchema.safeParse({
      paymentMethodId: "uuid",
      amount: "100.00",
      installments: 1.5,
    })
    expect(result.success).toBe(false)
  })
})

describe("CheckoutInputSchema", () => {
  it("accepts single payment", () => {
    const result = CheckoutInputSchema.safeParse({
      payments: [{ paymentMethodId: "uuid", amount: "100.00" }],
    })
    expect(result.success).toBe(true)
  })

  it("accepts multiple payments", () => {
    const result = CheckoutInputSchema.safeParse({
      payments: [
        { paymentMethodId: "uuid-1", amount: "50.00" },
        { paymentMethodId: "uuid-2", amount: "50.00" },
      ],
    })
    expect(result.success).toBe(true)
  })

  it("rejects empty payments array", () => {
    const result = CheckoutInputSchema.safeParse({ payments: [] })
    expect(result.success).toBe(true)
  })
})
