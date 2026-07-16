import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockProduct, mockProductList } from "../../__tests__/mocks"
import { 
  useUpdateProductMutation, 
  useActivateProductMutation, 
  useDeactivateProductMutation 
} from "../product.mutation"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useUpdateProductMutation", () => {
  it("updates a product successfully", async () => {
    const updatedProduct = { ...mockProduct, name: "Updated Name" }
    testServer.use(createMockHandler("put", `/products/${mockProduct.id}`, 200, updatedProduct))

    const { result } = renderHook(() => useUpdateProductMutation(), { wrapper: TestWrapper })

    result.current.mutate({ id: mockProduct.id, data: mockProduct })

    await waitFor(() => expect(result.current.isSuccess).toBe(true), { timeout: 5000 })
    expect(result.current.data?.name).toBe("Updated Name")
  })
})

describe("useActivateProductMutation", () => {
  it("activates a product", async () => {
    const activatedProduct = { ...mockProduct, isActive: true }
    testServer.use(createMockHandler("post", `/products/${mockProduct.id}/activate`, 200, activatedProduct))

    const { result } = renderHook(() => useActivateProductMutation(), { wrapper: TestWrapper })

    result.current.mutate(mockProduct.id)

    await waitFor(() => expect(result.current.isSuccess).toBe(true), { timeout: 5000 })
    expect(result.current.data?.isActive).toBe(true)
  })
})

describe("useDeactivateProductMutation", () => {
  it("deactivates a product", async () => {
    const deactivatedProduct = { ...mockProduct, isActive: false }
    testServer.use(createMockHandler("post", `/products/${mockProduct.id}/deactivate`, 200, deactivatedProduct))

    const { result } = renderHook(() => useDeactivateProductMutation(), { wrapper: TestWrapper })

    result.current.mutate(mockProduct.id)

    await waitFor(() => expect(result.current.isSuccess).toBe(true), { timeout: 5000 })
    expect(result.current.data?.isActive).toBe(false)
  })
})