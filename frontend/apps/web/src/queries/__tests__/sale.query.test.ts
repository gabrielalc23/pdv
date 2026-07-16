import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockSale, mockSaleList } from "../../__tests__/mocks"
import { useListSalesQuery, useGetSaleQuery } from "../sale.query"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useListSalesQuery", () => {
  it("fetches paginated sales", async () => {
    testServer.use(createMockHandler("get", "/sales", 200, mockSaleList))

    const { result } = renderHook(() => useListSalesQuery(), { wrapper: TestWrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.data).toHaveLength(1)
  })

  it("filters by status", async () => {
    testServer.use(createMockHandler("get", "/sales", 200, mockSaleList))

    const { result } = renderHook(() => useListSalesQuery({ status: "OPEN" }), {
      wrapper: TestWrapper,
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})

describe("useGetSaleQuery", () => {
  it("fetches single sale", async () => {
    testServer.use(createMockHandler("get", `/sales/${mockSale.id}`, 200, mockSale))

    const { result } = renderHook(() => useGetSaleQuery(mockSale.id), {
      wrapper: TestWrapper,
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.status).toBe("OPEN")
  })
})
