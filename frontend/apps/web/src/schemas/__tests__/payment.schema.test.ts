import { describe, it, expect } from "vitest";
import {
  PaymentMethodsResponseSchema,
  SalePaymentsResponseSchema,
} from "../payment.schema";

describe("PaymentMethodsResponseSchema", () => {
  it("accepts empty list", () => {
    const result = PaymentMethodsResponseSchema.safeParse({ data: [] });
    expect(result.success).toBe(true);
  });

  it("accepts payment methods", () => {
    const result = PaymentMethodsResponseSchema.safeParse({
      data: [
        {
          id: "uuid",
          code: "CASH",
          name: "Dinheiro",
          kind: "CASH",
          isActive: true,
          allowsChange: true,
          allowsInstallments: false,
          maxInstallments: 1,
          feePercentage: "0.0000",
          createdAt: "...",
          updatedAt: "...",
        },
      ],
    });
    expect(result.success).toBe(true);
  });
});

describe("SalePaymentsResponseSchema", () => {
  it("accepts empty payments", () => {
    const result = SalePaymentsResponseSchema.safeParse({ data: [] });
    expect(result.success).toBe(true);
  });

  it("accepts payment without optional fields", () => {
    const result = SalePaymentsResponseSchema.safeParse({
      data: [
        {
          id: "uuid",
          saleId: "uuid",
          paymentMethodId: "uuid",
          paymentMethodCode: "PIX",
          paymentMethodName: "Pix",
          amount: "100.00",
          status: "APPROVED",
          installments: 1,
          createdAt: "...",
          updatedAt: "...",
        },
      ],
    });
    expect(result.success).toBe(true);
  });
});
