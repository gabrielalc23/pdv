import type { PaginationMeta } from "./pagination.interface";

export interface ProductResponse {
  id: string;
  sku: string;
  barcode: string | null;
  name: string;
  categoryId?: string | null;
  categoryName?: string | null;
  price: string;
  cost: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface UpsertProductInput {
  sku: string;
  barcode: string | null;
  name: string;
  price: string;
  cost: string | null;
  categoryId?: string | null;
}

export interface ListProductsParams {
  search?: string;
  page?: number;
  pageSize?: number;
  activeOnly?: boolean;
  categoryId?: string;
}

export interface ProductListResponse {
  data: ProductResponse[];
  pagination: PaginationMeta;
}
