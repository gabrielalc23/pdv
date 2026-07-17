export interface PaymentMethodResponse {
  id: string;
  code: string;
  name: string;
  kind: string;
  isActive: boolean;
  allowsChange: boolean;
  allowsInstallments: boolean;
  maxInstallments: number;
  feePercentage: string;
  createdAt: string;
  updatedAt: string;
}

export interface PaymentMethodsResponse {
  data: PaymentMethodResponse[];
}

export interface SalePaymentResponse {
  id: string;
  saleId: string;
  paymentMethodId: string;
  paymentMethodCode: string;
  paymentMethodName: string;
  amount: string;
  receivedAmount?: string;
  changeAmount?: string;
  status: string;
  installments: number;
  externalReference?: string;
  paidAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SalePaymentsResponse {
  data: SalePaymentResponse[];
}
