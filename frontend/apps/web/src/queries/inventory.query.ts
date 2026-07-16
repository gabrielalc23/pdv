import { createApiCall, HttpMethod } from "@pdv/http"
import { z } from "zod"
import { mapApiError } from "@pdv/errors"
import { useQuery } from "@tanstack/react-query"
import type { UseQueryResult } from "@tanstack/react-query"
import {
  InventoryResponseSchema,
  InventoryListResponseSchema,
  ListInventoryParamsSchema,
  InventoryMovementListResponseSchema,
  ListInventoryMovementsParamsSchema,
} from "../schemas/inventory.schema"
import type {
  InventoryResponse,
  InventoryListResponse,
  InventoryMovementListResponse,
  ListInventoryParams,
  ListInventoryMovementsParams,
} from "../interfaces/inventory.interface"

function listInventory(params: ListInventoryParams = {}): Promise<InventoryListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/inventory",
    requestSchema: ListInventoryParamsSchema,
    responseSchema: InventoryListResponseSchema,
  })

  return api(params).catch(mapApiError)
}

function getProductInventory(productId: string): Promise<InventoryResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/products/${productId}/inventory`,
    requestSchema: z.object({}),
    responseSchema: InventoryResponseSchema,
  })

  return api({}).catch(mapApiError)
}

function listInventoryMovements(
  productId: string,
  params: ListInventoryMovementsParams = {},
): Promise<InventoryMovementListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/products/${productId}/inventory/movements`,
    requestSchema: ListInventoryMovementsParamsSchema,
    responseSchema: InventoryMovementListResponseSchema,
  })

  return api(params).catch(mapApiError)
}

export const inventoryKeys = {
  all: ["inventory"] as const,
  lists: () => [...inventoryKeys.all, "list"] as const,
  list: (params?: ListInventoryParams) => [...inventoryKeys.lists(), params] as const,
  details: () => [...inventoryKeys.all, "detail"] as const,
  detail: (productId: string) => [...inventoryKeys.details(), productId] as const,
  movements: (productId: string, params?: ListInventoryMovementsParams) =>
    [...inventoryKeys.all, "movements", productId, params] as const,
}

export function useListInventoryQuery(
  params?: ListInventoryParams,
): UseQueryResult<InventoryListResponse> {
  return useQuery({
    queryKey: inventoryKeys.list(params),
    queryFn: () => listInventory(params),
  })
}

export function useGetProductInventoryQuery(
  productId: string,
): UseQueryResult<InventoryResponse> {
  return useQuery({
    queryKey: inventoryKeys.detail(productId),
    queryFn: () => getProductInventory(productId),
    enabled: !!productId,
  })
}

export function useListInventoryMovementsQuery(
  productId: string,
  params?: ListInventoryMovementsParams,
): UseQueryResult<InventoryMovementListResponse> {
  return useQuery({
    queryKey: inventoryKeys.movements(productId, params),
    queryFn: () => listInventoryMovements(productId, params),
    enabled: !!productId,
  })
}
