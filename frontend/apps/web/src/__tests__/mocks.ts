import type { ProductResponse, ProductListResponse } from "../interfaces/product.interface"
import type { HealthResponse } from "../interfaces/health.interface"
import type { SaleResponse, SaleListResponse } from "../interfaces/sale.interface"
import type { PaymentMethodsResponse } from "../interfaces/payment.interface"
import type { CatalogProductResponse, CatalogListResponse } from "../interfaces/catalog.interface"
import type { InventoryResponse, InventoryMovementResponse } from "../interfaces/inventory.interface"
import type { FiscalDocumentResponse } from "../interfaces/fiscal.interface"
import type { ReceiptResponse } from "../interfaces/receipt.interface"

export const mockProduct: ProductResponse = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  sku: "ABC-123",
  barcode: null,
  name: "Produto Teste",
  price: "99.90",
  cost: null,
  isActive: true,
  createdAt: "2026-07-16T10:00:00Z",
  updatedAt: "2026-07-16T10:00:00Z",
}

export const mockProductList: ProductListResponse = {
  data: [mockProduct],
  pagination: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
}

export const mockHealth: HealthResponse = { status: "ok" }

export const mockSale: SaleResponse = {
  id: "550e8400-e29b-41d4-a716-446655440001",
  number: 1,
  status: "OPEN",
  subtotal: "100.00",
  discount: "0.00",
  addition: "0.00",
  total: "100.00",
  openedAt: "2026-07-16T10:00:00Z",
  completedAt: null,
  cancelledAt: null,
  createdAt: "2026-07-16T10:00:00Z",
  updatedAt: "2026-07-16T10:00:00Z",
  idempotencyKey: "key-001",
  items: [
    {
      id: "item-001",
      saleId: "550e8400-e29b-41d4-a716-446655440001",
      productId: "550e8400-e29b-41d4-a716-446655440000",
      productName: "Produto Teste",
      productSku: "ABC-123",
      unitPrice: "100.00",
      quantity: "1",
      discount: "0.00",
      total: "100.00",
      createdAt: "...",
    },
  ],
}

export const mockSaleList: SaleListResponse = {
  data: [mockSale],
  pagination: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
}

export const mockPaymentMethods: PaymentMethodsResponse = {
  data: [
    {
      id: "uuid-1",
      code: "CASH",
      name: "Dinheiro",
      kind: "CASH",
      isActive: true,
      allowsChange: true,
      allowsInstallments: false,
      maxInstallments: 1,
      feePercentage: "0.0000",
      createdAt: "...",
      updatedAt: "...",
    },
    {
      id: "uuid-2",
      code: "PIX",
      name: "Pix",
      kind: "PIX",
      isActive: true,
      allowsChange: false,
      allowsInstallments: false,
      maxInstallments: 1,
      feePercentage: "0.0000",
      createdAt: "...",
      updatedAt: "...",
    },
  ],
}

export const mockCatalogProduct: CatalogProductResponse = {
  id: "550e8400-e29b-41d4-a716-446655440002",
  sku: "SKU-CATALOG-001",
  barcode: "123456789012",
  name: "Produto Catalogo",
  price: "149.90",
  quantity: "5",
  isActive: true,
  inStock: true,
  createdAt: "...",
  updatedAt: "...",
}

export const mockCatalogProductList: CatalogListResponse = {
  data: [mockCatalogProduct],
  pagination: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
}

export const mockCatalogProductByBarcode: CatalogProductResponse = {
  id: "550e8400-e29b-41d4-a716-446655440003",
  sku: "SKU-BARCODE-001",
  barcode: "123456789012",
  name: "Produto Por Código de Barras",
  price: "99.90",
  quantity: "10",
  isActive: true,
  inStock: true,
  createdAt: "...",
  updatedAt: "...",
}

export const mockInventory: InventoryResponse = {
  productId: "550e8400-e29b-41d4-a716-446655440000",
  sku: "ABC-123",
  name: "Produto Estoque",
  quantity: "15",
  isActive: true,
  createdAt: "...",
  updatedAt: "...",
}

export const mockInventoryMovement: InventoryMovementResponse = {
  id: "550e8400-e29b-41d4-a716-446655440004",
  productId: "550e8400-e29b-41d4-a716-446655440000",
  type: "IN",
  quantity: "5",
  previousQuantity: "10",
  currentQuantity: "15",
  reason: "Restocking",
  referenceType: "PURCHASE",
  referenceId: "PO-001",
  createdAt: "...",
}

export const mockFiscalDocument: FiscalDocumentResponse = {
  id: "550e8400-e29b-41d4-a716-446655440005",
  saleId: "550e8400-e29b-41d4-a716-446655440001",
  status: "AUTHORIZED",
  environment: "PRODUCTION",
  documentModel: 65,
  series: 1,
  number: 123,
  accessKey: "35200600000000000000550000000000000000000000",
  protocol: "MOCK-123",
  provider: "mock",
  externalReference: "sale-001",
  xml: "<xml/>",
  errorCode: null,
  errorMessage: null,
  issuedAt: "...",
  cancelledAt: null,
  createdAt: "...",
  updatedAt: "...",
}

export const mockReceipt: ReceiptResponse = {
  sale: {
    id: "550e8400-e29b-41d4-a716-446655440001",
    number: 1,
    status: "COMPLETED",
    subtotal: "150.00",
    discount: "0.00",
    addition: "0.00",
    total: "150.00",
    openedAt: "...",
    completedAt: "...",
    cancelledAt: null,
    createdAt: "...",
    updatedAt: "...",
    idempotencyKey: "receipt-key-001",
  },
  items: [
    {
      productId: "550e8400-e29b-41d4-a716-446655440000",
      sku: "ABC-123",
      name: "Produto Teste",
      unitPrice: "100.00",
      quantity: "1",
      subtotal: "100.00",
      discount: "0.00",
      total: "100.00",
      createdAt: "...",
    },
  ],
  payments: [
    {
      method: "PIX",
      amount: "150.00",
      status: "APPROVED",
      installments: 1,
      receivedAmount: "150.00",
      changeAmount: null,
      externalReference: "pix-001",
    },
  ],
  fiscalDocument: mockFiscalDocument,
}
