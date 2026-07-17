import { z } from "zod";
import { PaginationMetaSchema } from "./pagination.schema";

export const ProductResponseSchema = z.object({
  id: z.string(),
  sku: z.string(),
  barcode: z.string().nullable(),
  name: z.string(),
  categoryId: z.string().nullable().optional(),
  categoryName: z.string().nullable().optional(),
  price: z.string(),
  cost: z.string().nullable(),
  isActive: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const UpsertProductInputSchema = z.object({
  sku: z.string(),
  barcode: z.string().nullable(),
  name: z.string(),
  price: z.string(),
  cost: z.string().nullable(),
  categoryId: z.string().nullable().optional(),
});

export const ListProductsParamsSchema = z.object({
  search: z.string().optional(),
  page: z.number().optional(),
  pageSize: z.number().optional(),
  activeOnly: z.boolean().optional(),
  categoryId: z.string().optional(),
});

export const ProductListResponseSchema = z.object({
  data: z.array(ProductResponseSchema),
  pagination: PaginationMetaSchema,
});
