import type { PaginationMeta } from "./pagination.interface";
import type { InventoryDirection } from "../types/inventory.type";

export interface InventoryResponse {
  productId: string;
  sku: string;
  barcode?: string;
  name: string;
  quantity: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ListInventoryParams {
  search?: string;
  page?: number;
  pageSize?: number;
  activeOnly?: boolean;
}

export interface InventoryListResponse {
  data: InventoryResponse[];
  pagination: PaginationMeta;
}

export interface CreateInventoryEntryInput {
  productId: string;
  quantity: string;
  reason?: string | null;
  referenceType: string;
  referenceId: string;
}

export interface CreateInventoryAdjustmentInput {
  productId: string;
  direction: InventoryDirection;
  quantity: string;
  reason: string;
  referenceType: string;
  referenceId: string;
}

export interface InventoryChangeSummary {
  productId: string;
  previousQuantity: string;
  currentQuantity: string;
  updatedAt: string;
}

export interface InventoryMovementResponse {
  id: string;
  productId: string;
  type: string;
  quantity: string;
  previousQuantity: string;
  currentQuantity: string;
  reason?: string;
  referenceType: string;
  referenceId: string;
  createdAt: string;
}

export interface InventoryChangeResponse {
  inventory: InventoryChangeSummary;
  movement: InventoryMovementResponse;
}

export interface ListInventoryMovementsParams {
  page?: number;
  pageSize?: number;
  type?: string;
}

export interface InventoryMovementListResponse {
  data: InventoryMovementResponse[];
  pagination: PaginationMeta;
}
