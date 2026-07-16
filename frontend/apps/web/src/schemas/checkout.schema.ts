import { z } from "zod"

export const CheckoutPaymentInputSchema = z.object({
  paymentMethodId: z.string(),
  amount: z.string(),
  receivedAmount: z.string().nullable().optional(),
  installments: z.number().int().nullable().optional(),
  externalReference: z.string().nullable().optional(),
})

export const CheckoutInputSchema = z.object({
  payments: z.array(CheckoutPaymentInputSchema),
})

export const CheckoutSaleResponseSchema = z.object({
  id: z.string(),
  number: z.number(),
  status: z.string(),
  subtotal: z.string(),
  discount: z.string(),
  addition: z.string(),
  total: z.string(),
  openedAt: z.string(),
  completedAt: z.string(),
  cancelledAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
  idempotencyKey: z.string(),
})

export const CheckoutPaymentResponseSchema = z.object({
  id: z.string(),
  saleId: z.string(),
  paymentMethodId: z.string(),
  paymentMethodCode: z.string(),
  paymentMethodName: z.string(),
  paymentMethodKind: z.string(),
  amount: z.string(),
  receivedAmount: z.string().optional(),
  changeAmount: z.string().optional(),
  status: z.string(),
  installments: z.number(),
  externalReference: z.string().optional(),
  paidAt: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const CheckoutFiscalDocumentResponseSchema = z.object({
  id: z.string(),
  saleId: z.string(),
  status: z.string(),
  environment: z.string(),
  documentModel: z.number(),
  series: z.number().optional(),
  number: z.number().optional(),
  accessKey: z.string().optional(),
  protocol: z.string().optional(),
  provider: z.string().optional(),
  externalReference: z.string().optional(),
  errorCode: z.string().optional(),
  errorMessage: z.string().optional(),
  issuedAt: z.string().optional(),
  cancelledAt: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const CheckoutResponseSchema = z.object({
  sale: CheckoutSaleResponseSchema,
  payments: z.array(CheckoutPaymentResponseSchema),
  fiscalDocument: CheckoutFiscalDocumentResponseSchema,
})
