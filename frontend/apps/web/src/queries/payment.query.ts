import { createApiCall, HttpMethod } from "@pdv/http"
import { z } from "zod"
import { mapApiError } from "@pdv/errors"
import { useQuery } from "@tanstack/react-query"
import type { UseQueryResult } from "@tanstack/react-query"
import {
  PaymentMethodsResponseSchema,
  SalePaymentsResponseSchema,
} from "../schemas/payment.schema"
import type {
  PaymentMethodsResponse,
  SalePaymentsResponse,
} from "../interfaces/payment.interface"

function listPaymentMethods(): Promise<PaymentMethodsResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/payment-methods",
    requestSchema: z.object({}),
    responseSchema: PaymentMethodsResponseSchema,
  })

  return api({}).catch(mapApiError)
}

function listSalePayments(saleId: string): Promise<SalePaymentsResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/sales/${saleId}/payments`,
    requestSchema: z.object({}),
    responseSchema: SalePaymentsResponseSchema,
  })

  return api({}).catch(mapApiError)
}

export const paymentKeys = {
  all: ["payments"] as const,
  methods: () => [...paymentKeys.all, "methods"] as const,
  salePayments: (saleId: string) => [...paymentKeys.all, "sale", saleId] as const,
}

export function useListPaymentMethodsQuery(): UseQueryResult<PaymentMethodsResponse> {
  return useQuery({
    queryKey: paymentKeys.methods(),
    queryFn: listPaymentMethods,
  })
}

export function useListSalePaymentsQuery(
  saleId: string,
): UseQueryResult<SalePaymentsResponse> {
  return useQuery({
    queryKey: paymentKeys.salePayments(saleId),
    queryFn: () => listSalePayments(saleId),
    enabled: !!saleId,
  })
}
