export interface ReceiptSaleResponse {
  id: string
  number: number
  status: string
  subtotal: string
  discount: string
  addition: string
  total: string
  openedAt: string
  completedAt: string
  cancelledAt: string | null
  createdAt: string
  updatedAt: string
  idempotencyKey: string
}

export interface ReceiptItemResponse {
  productId: string
  sku: string
  name: string
  unitPrice: string
  quantity: string
  subtotal: string
  discount: string
  total: string
  createdAt: string
}

export interface ReceiptPaymentResponse {
  method: string
  amount: string
  status: string
  installments: number
  receivedAmount?: string
  changeAmount?: string
  externalReference?: string
}

export interface ReceiptFiscalResponse {
  status: string
  accessKey?: string
  protocol?: string
  provider?: string
  externalReference?: string
  errorCode?: string
  errorMessage?: string
  issuedAt?: string
}

export interface ReceiptResponse {
  sale: ReceiptSaleResponse
  items: ReceiptItemResponse[]
  payments: ReceiptPaymentResponse[]
  fiscalDocument: ReceiptFiscalResponse | null
}
