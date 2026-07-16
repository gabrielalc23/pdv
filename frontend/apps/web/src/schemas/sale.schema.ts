import { z } from "zod"
import { PaginationMetaSchema } from "./pagination.schema"

export const SaleItemResponseSchema = z.object({
  id: z.string(),
  saleId: z.string(),
  productId: z.string(),
  productName: z.string(),
  productSku: z.string(),
  unitPrice: z.string(),
  quantity: z.string(),
  discount: z.string(),
  total: z.string(),
  createdAt: z.string(),
})

export const SaleResponseSchema = z.object({
  id: z.string(),
  number: z.number(),
  status: z.string(),
  subtotal: z.string(),
  discount: z.string(),
  addition: z.string(),
  total: z.string(),
  openedAt: z.string(),
  completedAt: z.string().nullable(),
  cancelledAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
  idempotencyKey: z.string(),
  items: z.array(SaleItemResponseSchema),
})

export const SaleListItemResponseSchema = z.object({
  id: z.string(),
  number: z.number(),
  status: z.string(),
  subtotal: z.string(),
  discount: z.string(),
  addition: z.string(),
  total: z.string(),
  openedAt: z.string(),
  completedAt: z.string().nullable(),
  cancelledAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
  idempotencyKey: z.string(),
})

export const CreateSaleInputSchema = z.object({
  idempotencyKey: z.string(),
})

export const ListSalesParamsSchema = z.object({
  status: z.enum(["OPEN", "COMPLETED", "CANCELLED"]).optional(),
  page: z.number().optional(),
  pageSize: z.number().optional(),
})

export const AddSaleItemInputSchema = z.object({
  productId: z.string(),
  quantity: z.string(),
  discount: z.string().nullable().optional(),
})

export const UpdateSaleItemInputSchema = z.object({
  quantity: z.string(),
  discount: z.string().nullable().optional(),
})

export const SaleListResponseSchema = z.object({
  data: z.array(SaleListItemResponseSchema),
  pagination: PaginationMetaSchema,
})
