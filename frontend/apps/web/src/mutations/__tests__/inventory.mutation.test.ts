import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockInventoryMovement } from "../../__tests__/mocks"
import {
  useCreateInventoryEntryMutation,
  useCreateInventoryAdjustmentMutation,
} from "../inventory.mutation"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useCreateInventoryEntryMutation", () => {
  it("creates an inventory entry", async () => {
    const inventoryChangeResponse = {
      inventory: {
        productId: mockInventoryMovement.productId,
        previousQuantity: mockInventoryMovement.previousQuantity,
        currentQuantity: "20",
        updatedAt: "...",
      },
      movement: { ...mockInventoryMovement, currentQuantity: "20" },
    }
    testServer.use(createMockHandler("post", "/inventory/entries", 200, inventoryChangeResponse))

    const { result } = renderHook(() => useCreateInventoryEntryMutation(), { wrapper: TestWrapper })

    result.current.mutate({
      productId: mockInventoryMovement.productId,
      quantity: "20",
      reason: "Test entry",
      referenceType: "TEST",
      referenceId: "REF-123",
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.movement.currentQuantity).toBe("20")
  })
})

describe("useCreateInventoryAdjustmentMutation", () => {
  it("creates an inventory adjustment IN", async () => {
    const inventoryChangeResponse = {
      inventory: {
        productId: mockInventoryMovement.productId,
        previousQuantity: mockInventoryMovement.previousQuantity,
        currentQuantity: "25",
        updatedAt: "...",
      },
      movement: { ...mockInventoryMovement, currentQuantity: "25", type: "IN" },
    }
    testServer.use(createMockHandler("post", "/inventory/adjustments", 200, inventoryChangeResponse))

    const { result } = renderHook(() => useCreateInventoryAdjustmentMutation(), { wrapper: TestWrapper })

    result.current.mutate({
      productId: mockInventoryMovement.productId,
      direction: "IN",
      quantity: "5",
      reason: "Test IN",
      referenceType: "TEST",
      referenceId: "REF-456",
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.movement.currentQuantity).toBe("25")
    expect(result.current.data?.movement.type).toBe("IN")
  })

  it("creates an inventory adjustment OUT", async () => {
    const inventoryChangeResponse = {
      inventory: {
        productId: mockInventoryMovement.productId,
        previousQuantity: mockInventoryMovement.previousQuantity,
        currentQuantity: "5",
        updatedAt: "...",
      },
      movement: { ...mockInventoryMovement, currentQuantity: "5", type: "OUT" },
    }
    testServer.use(createMockHandler("post", "/inventory/adjustments", 200, inventoryChangeResponse))

    const { result } = renderHook(() => useCreateInventoryAdjustmentMutation(), { wrapper: TestWrapper })

    result.current.mutate({
      productId: mockInventoryMovement.productId,
      direction: "OUT",
      quantity: "10",
      reason: "Test OUT",
      referenceType: "TEST",
      referenceId: "REF-789",
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.movement.currentQuantity).toBe("5")
    expect(result.current.data?.movement.type).toBe("OUT")
  })
})