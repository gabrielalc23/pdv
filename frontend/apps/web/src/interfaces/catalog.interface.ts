import type { PaginationMeta } from "./pagination.interface";

export interface CatalogProductResponse {
  id: string;
  sku: string;
  barcode: string | null;
  name: string;
  categoryId?: string | null;
  categoryName?: string | null;
  price: string;
  quantity: string;
  isActive: boolean;
  inStock: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ListCatalogParams {
  search?: string;
  page?: number;
  pageSize?: number;
  activeOnly?: boolean;
  inStockOnly?: boolean;
  categoryId?: string;
}

export interface CatalogListResponse {
  data: CatalogProductResponse[];
  pagination: PaginationMeta;
}
