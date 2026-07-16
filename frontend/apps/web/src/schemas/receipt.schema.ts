import { z } from "zod"

export const ReceiptSaleResponseSchema = z.object({
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

export const ReceiptItemResponseSchema = z.object({
  productId: z.string(),
  sku: z.string(),
  name: z.string(),
  unitPrice: z.string(),
  quantity: z.string(),
  subtotal: z.string(),
  discount: z.string(),
  total: z.string(),
  createdAt: z.string(),
})

export const ReceiptPaymentResponseSchema = z.object({
  method: z.string(),
  amount: z.string(),
  status: z.string(),
  installments: z.number(),
  receivedAmount: z.string().optional(),
  changeAmount: z.string().optional(),
  externalReference: z.string().optional(),
})

export const ReceiptFiscalResponseSchema = z.object({
  status: z.string(),
  accessKey: z.string().optional(),
  protocol: z.string().optional(),
  provider: z.string().optional(),
  externalReference: z.string().optional(),
  errorCode: z.string().optional(),
  errorMessage: z.string().optional(),
  issuedAt: z.string().optional(),
})

export const ReceiptResponseSchema = z.object({
  sale: ReceiptSaleResponseSchema,
  items: z.array(ReceiptItemResponseSchema),
  payments: z.array(ReceiptPaymentResponseSchema),
  fiscalDocument: ReceiptFiscalResponseSchema.nullable(),
})
