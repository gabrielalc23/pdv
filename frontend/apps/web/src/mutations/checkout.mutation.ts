import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { UseMutationResult } from "@tanstack/react-query"
import { createApiCall, HttpMethod } from "@pdv/http"
import { mapApiError } from "@pdv/errors"
import {
  CheckoutInputSchema,
  CheckoutResponseSchema,
} from "../schemas/checkout.schema"
import type {
  CheckoutInput,
  CheckoutResponse,
} from "../interfaces/checkout.interface"
import { saleKeys } from "../queries/sale.query"

function checkoutSale(saleId: string, data: CheckoutInput): Promise<CheckoutResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: `/sales/${saleId}/checkout`,
    requestSchema: CheckoutInputSchema,
    responseSchema: CheckoutResponseSchema,
  })

  return api(data).catch(mapApiError)
}

export function useCheckoutSaleMutation(): UseMutationResult<
  CheckoutResponse,
  Error,
  { saleId: string; data: CheckoutInput }
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ saleId, data }) => checkoutSale(saleId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: saleKeys.detail(result.sale.id) })
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}
