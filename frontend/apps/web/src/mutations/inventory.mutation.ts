import { useMutation, useQueryClient } from "@tanstack/react-query"
import type { QueryClient, UseMutationResult } from "@tanstack/react-query"
import { createApiCall, HttpMethod } from "@pdv/http"
import { mapApiError } from "@pdv/errors"
import {
  CreateInventoryEntryInputSchema,
  CreateInventoryAdjustmentInputSchema,
  InventoryChangeResponseSchema,
} from "../schemas/inventory.schema"
import type {
  InventoryChangeResponse,
  CreateInventoryEntryInput,
  CreateInventoryAdjustmentInput,
} from "../interfaces/inventory.interface"
import { inventoryKeys } from "../queries/inventory.query"

function createInventoryEntry(
  data: CreateInventoryEntryInput,
): Promise<InventoryChangeResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: "/inventory/entries",
    requestSchema: CreateInventoryEntryInputSchema,
    responseSchema: InventoryChangeResponseSchema,
  })

  return api(data).catch(mapApiError)
}

function createInventoryAdjustment(
  data: CreateInventoryAdjustmentInput,
): Promise<InventoryChangeResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: "/inventory/adjustments",
    requestSchema: CreateInventoryAdjustmentInputSchema,
    responseSchema: InventoryChangeResponseSchema,
  })

  return api(data).catch(mapApiError)
}

export function useCreateInventoryEntryMutation(): UseMutationResult<
  InventoryChangeResponse,
  Error,
  CreateInventoryEntryInput
> {
  const queryClient: QueryClient = useQueryClient()

  return useMutation({
    mutationFn: createInventoryEntry,
    onSuccess: (result: InventoryChangeResponse): void => {
      queryClient.invalidateQueries({
        queryKey: inventoryKeys.detail(result.movement.productId),
      })
      queryClient.invalidateQueries({ queryKey: inventoryKeys.lists() })
      queryClient.invalidateQueries({
        queryKey: inventoryKeys.movements(result.movement.productId),
      })
    },
  })
}

export function useCreateInventoryAdjustmentMutation(): UseMutationResult<
  InventoryChangeResponse,
  Error,
  CreateInventoryAdjustmentInput
> {
  const queryClient: QueryClient = useQueryClient()

  return useMutation({
    mutationFn: createInventoryAdjustment,
    onSuccess: (result: InventoryChangeResponse): void => {
      queryClient.invalidateQueries({
        queryKey: inventoryKeys.detail(result.movement.productId),
      })
      queryClient.invalidateQueries({ queryKey: inventoryKeys.lists() })
      queryClient.invalidateQueries({
        queryKey: inventoryKeys.movements(result.movement.productId),
      })
    },
  })
}
