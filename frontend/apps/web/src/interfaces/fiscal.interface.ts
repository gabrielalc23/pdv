import type { Nullable } from "@pdv/types"

export interface FiscalDocumentResponse {
  id: string
  saleId: string
  status: string
  environment: string
  documentModel: number
  series?: number
  number?: number
  accessKey?: string
  protocol?: string
  provider?: string
  externalReference?: string
  xml?: string
  errorCode?: Nullable<string>
  errorMessage?: Nullable<string>
  issuedAt?: Nullable<string>
  cancelledAt?: Nullable<string>
  createdAt: string
  updatedAt: string
}
