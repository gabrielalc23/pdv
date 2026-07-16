import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockProduct, mockProductList } from "../../__tests__/mocks"
import { useListProductsQuery, useGetProductQuery } from "../product.query"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useListProductsQuery", () => {
  it("fetches paginated products", async () => {
    const productUrl = "/products"
    testServer.use(createMockHandler("get", productUrl, 200, mockProductList))

    const { result } = renderHook(() => useListProductsQuery(), { wrapper: TestWrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.data).toHaveLength(1)
    expect(result.current.data?.pagination.total).toBe(1)
  })

  it("accepts search params", async () => {
    const productUrl = "/products"
    testServer.use(createMockHandler("get", productUrl, 200, mockProductList))

    const { result } = renderHook(() => useListProductsQuery({ search: "ABC" }), {
      wrapper: TestWrapper,
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})

describe("useGetProductQuery", () => {
  it("fetches single product", async () => {
    const productId = mockProduct.id
    const productUrl = `/products/${productId}`
    testServer.use(createMockHandler("get", productUrl, 200, mockProduct))

    const { result } = renderHook(() => useGetProductQuery(productId), {
      wrapper: TestWrapper,
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.name).toBe("Produto Teste")
  })

  it("is disabled when id is empty", () => {
    const { result } = renderHook(() => useGetProductQuery(""), { wrapper: TestWrapper })

    expect(result.current.fetchStatus).toBe("idle")
  })
})
