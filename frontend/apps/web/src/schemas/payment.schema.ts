import { z } from "zod"

export const PaymentMethodResponseSchema = z.object({
  id: z.string(),
  code: z.string(),
  name: z.string(),
  kind: z.string(),
  isActive: z.boolean(),
  allowsChange: z.boolean(),
  allowsInstallments: z.boolean(),
  maxInstallments: z.number(),
  feePercentage: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const PaymentMethodsResponseSchema = z.object({
  data: z.array(PaymentMethodResponseSchema),
})

export const SalePaymentResponseSchema = z.object({
  id: z.string(),
  saleId: z.string(),
  paymentMethodId: z.string(),
  paymentMethodCode: z.string(),
  paymentMethodName: z.string(),
  amount: z.string(),
  receivedAmount: z.string().optional(),
  changeAmount: z.string().optional(),
  status: z.string(),
  installments: z.number(),
  externalReference: z.string().optional(),
  paidAt: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const SalePaymentsResponseSchema = z.object({
  data: z.array(SalePaymentResponseSchema),
})
