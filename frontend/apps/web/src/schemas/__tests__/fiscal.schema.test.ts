import { describe, it, expect } from "vitest"
import { FiscalDocumentResponseSchema } from "../fiscal.schema"

describe("FiscalDocumentResponseSchema", () => {
  it("accepts minimal fiscal document", () => {
    const result = FiscalDocumentResponseSchema.safeParse({
      id: "uuid",
      saleId: "uuid",
      status: "PENDING",
      environment: "HOMOLOGATION",
      documentModel: 65,
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(true)
  })

  it("accepts authorized document with all fields", () => {
    const result = FiscalDocumentResponseSchema.safeParse({
      id: "uuid",
      saleId: "uuid",
      status: "AUTHORIZED",
      environment: "PRODUCTION",
      documentModel: 65,
      series: 1,
      number: 123,
      accessKey: "35200600000000000000550000000000000000000000",
      protocol: "MOCK-123",
      provider: "mock",
      externalReference: "sale-uuid",
      xml: "<xml/>",
      issuedAt: "2026-07-16T10:00:00Z",
      createdAt: "...",
      updatedAt: "...",
    })
    expect(result.success).toBe(true)
  })
})
