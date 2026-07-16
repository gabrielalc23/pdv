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
  errorCode?: string
  errorMessage?: string
  issuedAt?: string
  cancelledAt?: string
  createdAt: string
  updatedAt: string
}
