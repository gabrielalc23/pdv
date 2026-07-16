import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockSale } from "../../__tests__/mocks"
import {
  useAddSaleItemMutation,
  useUpdateSaleItemMutation,
  useRemoveSaleItemMutation,
  useCancelSaleMutation,
} from "../sale.mutation"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useAddSaleItemMutation", () => {
  it("adds an item to a sale", async () => {
    const saleWithItem = {
      ...mockSale,
      items: [{ ...mockSale.items[0], id: "item-123" }],
    }
    testServer.use(createMockHandler("post", `/sales/${mockSale.id}/items`, 200, saleWithItem))

    const { result } = renderHook(() => useAddSaleItemMutation(), { wrapper: TestWrapper })

    result.current.mutate({ saleId: mockSale.id, data: mockSale.items[0] })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.items).toHaveLength(1)
    expect(result.current.data?.items[0].id).toBe("item-123")
  })
})

describe("useUpdateSaleItemMutation", () => {
  it("updates a sale item", async () => {
    const saleWithUpdatedItem = {
      ...mockSale,
      items: [{ ...mockSale.items[0], quantity: "5" }],
    }
    testServer.use(createMockHandler("put", `/sales/${mockSale.id}/items/${mockSale.items[0].id}`, 200, saleWithUpdatedItem))

    const { result } = renderHook(() => useUpdateSaleItemMutation(), { wrapper: TestWrapper })

    result.current.mutate({
      saleId: mockSale.id,
      itemId: mockSale.items[0].id,
      data: { quantity: "5" },
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.items[0].quantity).toBe("5")
  })
})

describe("useRemoveSaleItemMutation", () => {
  it("removes an item from a sale", async () => {
    const saleWithoutItem = {
      ...mockSale,
      items: [],
    }
    testServer.use(createMockHandler("delete", `/sales/${mockSale.id}/items/${mockSale.items[0].id}`, 200, saleWithoutItem))

    const { result } = renderHook(() => useRemoveSaleItemMutation(), { wrapper: TestWrapper })

    result.current.mutate({ saleId: mockSale.id, itemId: mockSale.items[0].id })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.items).toHaveLength(0)
  })
})

describe("useCancelSaleMutation", () => {
  it("cancels a sale", async () => {
    const cancelledSale = {
      ...mockSale,
      status: "CANCELLED",
      cancelledAt: "2026-07-16T10:00:00Z",
    }
    testServer.use(createMockHandler("post", `/sales/${mockSale.id}/cancel`, 200, cancelledSale))

    const { result } = renderHook(() => useCancelSaleMutation(), { wrapper: TestWrapper })

    result.current.mutate(mockSale.id)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.status).toBe("CANCELLED")
    expect(result.current.data?.cancelledAt).toBe("2026-07-16T10:00:00Z")
  })
})