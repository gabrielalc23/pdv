import type { PaginationMeta } from "./pagination.interface"

export interface ProductResponse {
  id: string
  sku: string
  barcode: string | null
  name: string
  price: string
  cost: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface UpsertProductInput {
  sku: string
  barcode: string | null
  name: string
  price: string
  cost: string | null
}

export interface ListProductsParams {
  search?: string
  page?: number
  pageSize?: number
  activeOnly?: boolean
}

export interface ProductListResponse {
  data: ProductResponse[]
  pagination: PaginationMeta
}
