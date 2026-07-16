import type { PaginationMeta } from "./pagination.interface"
import type { SaleStatus } from "../types/sale.type"

export interface SaleItemResponse {
  id: string
  saleId: string
  productId: string
  productName: string
  productSku: string
  unitPrice: string
  quantity: string
  discount: string
  total: string
  createdAt: string
}

export interface SaleResponse {
  id: string
  number: number
  status: string
  subtotal: string
  discount: string
  addition: string
  total: string
  openedAt: string
  completedAt: string | null
  cancelledAt: string | null
  createdAt: string
  updatedAt: string
  idempotencyKey: string
  items: SaleItemResponse[]
}

export interface SaleListItemResponse {
  id: string
  number: number
  status: string
  subtotal: string
  discount: string
  addition: string
  total: string
  openedAt: string
  completedAt: string | null
  cancelledAt: string | null
  createdAt: string
  updatedAt: string
  idempotencyKey: string
}

export interface CreateSaleInput {
  idempotencyKey: string
}

export interface ListSalesParams {
  status?: SaleStatus
  page?: number
  pageSize?: number
}

export interface AddSaleItemInput {
  productId: string
  quantity: string
  discount?: string | null
}

export interface UpdateSaleItemInput {
  quantity: string
  discount?: string | null
}

export interface SaleListResponse {
  data: SaleListItemResponse[]
  pagination: PaginationMeta
}
