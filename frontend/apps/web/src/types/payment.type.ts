export type PaymentMethodKind =
  | "CASH"
  | "PIX"
  | "DEBIT_CARD"
  | "CREDIT_CARD"
  | "VOUCHER"
  | "STORE_CREDIT"
  | "OTHER"

export type PaymentStatus = "PENDING" | "APPROVED" | "DECLINED" | "CANCELLED"
