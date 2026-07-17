import type { Nullable } from "@pdv/types";

export interface ReceiptSaleResponse {
  id: string;
  number: number;
  status: string;
  subtotal: string;
  discount: string;
  addition: string;
  total: string;
  openedAt: string;
  completedAt: string;
  cancelledAt: string | null;
  createdAt: string;
  updatedAt: string;
  idempotencyKey: string;
}

export interface ReceiptItemResponse {
  productId: string;
  sku: string;
  name: string;
  unitPrice: string;
  quantity: string;
  subtotal: string;
  discount: string;
  total: string;
  createdAt: string;
}

export interface ReceiptPaymentResponse {
  method: string;
  amount: string;
  status: string;
  installments: number;
  receivedAmount?: Nullable<string>;
  changeAmount?: Nullable<string>;
  externalReference?: Nullable<string>;
}

export interface ReceiptFiscalResponse {
  status: string;
  accessKey?: Nullable<string>;
  protocol?: Nullable<string>;
  provider?: Nullable<string>;
  externalReference?: Nullable<string>;
  errorCode?: Nullable<string>;
  errorMessage?: Nullable<string>;
  issuedAt?: Nullable<string>;
}

export interface ReceiptResponse {
  sale: ReceiptSaleResponse;
  items: ReceiptItemResponse[];
  payments: ReceiptPaymentResponse[];
  fiscalDocument: Nullable<ReceiptFiscalResponse>;
}
