import { z } from "zod";
import { PaginationMetaSchema } from "./pagination.schema";

export const CatalogProductResponseSchema = z.object({
  id: z.string(),
  sku: z.string(),
  barcode: z.string().nullable(),
  name: z.string(),
  categoryId: z.string().nullable().optional(),
  categoryName: z.string().nullable().optional(),
  price: z.string(),
  quantity: z.string(),
  isActive: z.boolean(),
  inStock: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const ListCatalogParamsSchema = z.object({
  search: z.string().optional(),
  page: z.number().optional(),
  pageSize: z.number().optional(),
  activeOnly: z.boolean().optional(),
  inStockOnly: z.boolean().optional(),
  categoryId: z.string().optional(),
});

export const CatalogListResponseSchema = z.object({
  data: z.array(CatalogProductResponseSchema),
  pagination: PaginationMetaSchema,
});
