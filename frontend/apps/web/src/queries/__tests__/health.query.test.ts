import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createMockHandler, testServer, TestWrapper } from "../../__tests__/test-utils"
import { mockHealth } from "../../__tests__/mocks"
import { useHealthQuery } from "../health.query"

beforeAll(() => testServer.listen())
afterAll(() => testServer.close())
afterEach(() => testServer.resetHandlers())

describe("useHealthQuery", () => {
  it("returns health status", async () => {
    testServer.use(createMockHandler("get", "/health", 200, mockHealth))

    const { result } = renderHook(() => useHealthQuery(), { wrapper: TestWrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockHealth)
  })
})
