import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import {
  createMockHandler,
  testServer,
  TestWrapper,
} from "../../__tests__/test-utils";
import { mockPaymentMethods } from "../../__tests__/mocks";
import { useListPaymentMethodsQuery } from "../payment.query";

beforeAll(() => testServer.listen());
afterAll(() => testServer.close());
afterEach(() => testServer.resetHandlers());

describe("useListPaymentMethodsQuery", () => {
  it("fetches payment methods", async () => {
    testServer.use(
      createMockHandler("get", "/payment-methods", 200, mockPaymentMethods),
    );

    const { result } = renderHook(() => useListPaymentMethodsQuery(), {
      wrapper: TestWrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(2);
  });
});
