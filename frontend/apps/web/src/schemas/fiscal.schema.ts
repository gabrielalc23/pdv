import { z } from "zod"

export const FiscalDocumentResponseSchema = z.object({
  id: z.string(),
  saleId: z.string(),
  status: z.string(),
  environment: z.string(),
  documentModel: z.number(),
  series: z.number().optional(),
  number: z.number().optional(),
  accessKey: z.string().optional(),
  protocol: z.string().optional(),
  provider: z.string().optional(),
  externalReference: z.string().optional(),
  xml: z.string().optional(),
  errorCode: z.string().optional(),
  errorMessage: z.string().optional(),
  issuedAt: z.string().optional(),
  cancelledAt: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
})
