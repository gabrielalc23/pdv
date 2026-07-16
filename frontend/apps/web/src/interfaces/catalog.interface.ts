import type { PaginationMeta } from "./pagination.interface"

export interface CatalogProductResponse {
  id: string
  sku: string
  barcode: string | null
  name: string
  price: string
  quantity: string
  isActive: boolean
  inStock: boolean
  createdAt: string
  updatedAt: string
}

export interface ListCatalogParams {
  search?: string
  page?: number
  pageSize?: number
  activeOnly?: boolean
  inStockOnly?: boolean
}

export interface CatalogListResponse {
  data: CatalogProductResponse[]
  pagination: PaginationMeta
}
