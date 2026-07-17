import { z } from "zod";
import { PaginationMetaSchema } from "./pagination.schema";

export const InventoryResponseSchema = z.object({
  productId: z.string(),
  sku: z.string(),
  barcode: z.string().optional(),
  name: z.string(),
  quantity: z.string(),
  isActive: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const ListInventoryParamsSchema = z.object({
  search: z.string().optional(),
  page: z.number().optional(),
  pageSize: z.number().optional(),
  activeOnly: z.boolean().optional(),
});

export const InventoryListResponseSchema = z.object({
  data: z.array(InventoryResponseSchema),
  pagination: PaginationMetaSchema,
});

export const CreateInventoryEntryInputSchema = z.object({
  productId: z.string(),
  quantity: z.string(),
  reason: z.string().nullable().optional(),
  referenceType: z.string(),
  referenceId: z.string(),
});

export const CreateInventoryAdjustmentInputSchema = z.object({
  productId: z.string(),
  direction: z.enum(["IN", "OUT"]),
  quantity: z.string(),
  reason: z.string(),
  referenceType: z.string(),
  referenceId: z.string(),
});

export const InventoryChangeSummarySchema = z.object({
  productId: z.string(),
  previousQuantity: z.string(),
  currentQuantity: z.string(),
  updatedAt: z.string(),
});

export const InventoryMovementResponseSchema = z.object({
  id: z.string(),
  productId: z.string(),
  type: z.string(),
  quantity: z.string(),
  previousQuantity: z.string(),
  currentQuantity: z.string(),
  reason: z.string().optional(),
  referenceType: z.string(),
  referenceId: z.string(),
  createdAt: z.string(),
});

export const InventoryChangeResponseSchema = z.object({
  inventory: InventoryChangeSummarySchema,
  movement: InventoryMovementResponseSchema,
});

export const ListInventoryMovementsParamsSchema = z.object({
  page: z.number().optional(),
  pageSize: z.number().optional(),
  type: z.string().optional(),
});

export const InventoryMovementListResponseSchema = z.object({
  data: z.array(InventoryMovementResponseSchema),
  pagination: PaginationMetaSchema,
});
