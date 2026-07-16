import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { UseMutationResult } from "@tanstack/react-query"
import { createApiCall, HttpMethod } from "@pdv/http"
import { z } from "zod"
import { mapApiError } from "@pdv/errors"
import {
  SaleResponseSchema,
  CreateSaleInputSchema,
  AddSaleItemInputSchema,
  UpdateSaleItemInputSchema,
} from "../schemas/sale.schema"
import type {
  SaleResponse,
  CreateSaleInput,
  AddSaleItemInput,
  UpdateSaleItemInput,
} from "../interfaces/sale.interface"
import { saleKeys } from "../queries/sale.query"

function createSale(data: CreateSaleInput): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: "/sales",
    requestSchema: CreateSaleInputSchema,
    responseSchema: SaleResponseSchema,
  })

  return api(data).catch(mapApiError)
}

function addSaleItem(saleId: string, data: AddSaleItemInput): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: `/sales/${saleId}/items`,
    requestSchema: AddSaleItemInputSchema,
    responseSchema: SaleResponseSchema,
  })

  return api(data).catch(mapApiError)
}

function updateSaleItem(
  saleId: string,
  itemId: string,
  data: UpdateSaleItemInput,
): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.PUT,
    path: `/sales/${saleId}/items/${itemId}`,
    requestSchema: UpdateSaleItemInputSchema,
    responseSchema: SaleResponseSchema,
  })

  return api(data).catch(mapApiError)
}

function removeSaleItem(saleId: string, itemId: string): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.DELETE,
    path: `/sales/${saleId}/items/${itemId}`,
    requestSchema: z.object({}),
    responseSchema: SaleResponseSchema,
  })

  return api({}).catch(mapApiError)
}

function cancelSale(id: string): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: `/sales/${id}/cancel`,
    requestSchema: z.object({}),
    responseSchema: SaleResponseSchema,
  })

  return api({}).catch(mapApiError)
}

export function useCreateSaleMutation(): UseMutationResult<
  SaleResponse,
  Error,
  CreateSaleInput
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createSale,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}

export function useAddSaleItemMutation(): UseMutationResult<
  SaleResponse,
  Error,
  { saleId: string; data: AddSaleItemInput }
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ saleId, data }) => addSaleItem(saleId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: saleKeys.detail(result.id) })
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}

export function useUpdateSaleItemMutation(): UseMutationResult<
  SaleResponse,
  Error,
  { saleId: string; itemId: string; data: UpdateSaleItemInput }
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ saleId, itemId, data }) => updateSaleItem(saleId, itemId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: saleKeys.detail(result.id) })
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}

export function useRemoveSaleItemMutation(): UseMutationResult<
  SaleResponse,
  Error,
  { saleId: string; itemId: string }
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ saleId, itemId }) => removeSaleItem(saleId, itemId),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: saleKeys.detail(result.id) })
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}

export function useCancelSaleMutation(): UseMutationResult<
  SaleResponse,
  Error,
  string
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: cancelSale,
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: saleKeys.detail(result.id) })
      queryClient.invalidateQueries({ queryKey: saleKeys.lists() })
    },
  })
}
