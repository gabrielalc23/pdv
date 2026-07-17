import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import {
  createMockHandler,
  testServer,
  TestWrapper,
} from "../../__tests__/test-utils";
import { mockSale } from "../../__tests__/mocks";
import { useCheckoutSaleMutation } from "../checkout.mutation";

beforeAll(() => testServer.listen());
afterAll(() => testServer.close());
afterEach(() => testServer.resetHandlers());

describe("useCheckoutSaleMutation", () => {
  it("completes checkout with single payment", async () => {
    const checkoutResult = {
      sale: {
        ...mockSale,
        status: "COMPLETED",
        completedAt: "2026-07-16T10:00:00Z",
      },
      payments: [
        {
          id: "pay-123",
          saleId: mockSale.id,
          paymentMethodId: "pm-001",
          paymentMethodCode: "PIX",
          paymentMethodName: "Pix",
          paymentMethodKind: "PIX",
          amount: "100.00",
          receivedAmount: "100.00",
          changeAmount: "0.00",
          status: "APPROVED",
          installments: 1,
          externalReference: "checkout-001",
          paidAt: "2026-07-16T10:00:00Z",
          createdAt: "...",
          updatedAt: "...",
        },
      ],
      fiscalDocument: {
        id: "doc-123",
        saleId: mockSale.id,
        status: "AUTHORIZED",
        environment: "PRODUCTION",
        documentModel: 65,
        series: 1,
        number: 123,
        accessKey: "35200600000000000000550000000000000000000000",
        protocol: "MOCK-123",
        provider: "mock",
        externalReference: "checkout-001",
        xml: "<xml/>",
        errorCode: "",
        errorMessage: "",
        issuedAt: "...",
        cancelledAt: "",
        createdAt: "...",
        updatedAt: "...",
      },
    };
    testServer.use(
      createMockHandler(
        "post",
        `/sales/${mockSale.id}/checkout`,
        200,
        checkoutResult,
      ),
    );

    const { result } = renderHook(() => useCheckoutSaleMutation(), {
      wrapper: TestWrapper,
    });

    result.current.mutate({
      saleId: mockSale.id,
      data: {
        payments: [
          {
            paymentMethodId: "pm-001",
            amount: "100.00",
            receivedAmount: "100.00",
            installments: 1,
            externalReference: "checkout-001",
          },
        ],
      },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.sale.status).toBe("COMPLETED");
    expect(result.current.data?.sale.completedAt).toBe("2026-07-16T10:00:00Z");
    expect(result.current.data?.payments).toHaveLength(1);
    expect(result.current.data?.payments[0].status).toBe("APPROVED");
  });
});
