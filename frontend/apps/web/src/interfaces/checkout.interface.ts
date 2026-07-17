export interface CheckoutPaymentInput {
  paymentMethodId: string;
  amount: string;
  receivedAmount?: string | null;
  installments?: number | null;
  externalReference?: string | null;
}

export interface CheckoutInput {
  payments: CheckoutPaymentInput[];
}

export interface CheckoutSaleResponse {
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

export interface CheckoutPaymentResponse {
  id: string;
  saleId: string;
  paymentMethodId: string;
  paymentMethodCode: string;
  paymentMethodName: string;
  paymentMethodKind: string;
  amount: string;
  receivedAmount?: string;
  changeAmount?: string;
  status: string;
  installments: number;
  externalReference?: string;
  paidAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface CheckoutFiscalDocumentResponse {
  id: string;
  saleId: string;
  status: string;
  environment: string;
  documentModel: number;
  series?: number;
  number?: number;
  accessKey?: string;
  protocol?: string;
  provider?: string;
  externalReference?: string;
  errorCode?: string;
  errorMessage?: string;
  issuedAt?: string;
  cancelledAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CheckoutResponse {
  sale: CheckoutSaleResponse;
  payments: CheckoutPaymentResponse[];
  fiscalDocument: CheckoutFiscalDocumentResponse;
}
