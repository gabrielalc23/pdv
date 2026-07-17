import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { QueryClient, UseMutationResult } from "@tanstack/react-query";
import { createApiCall, HttpMethod } from "@pdv/http";
import { z } from "zod";
import { mapApiError } from "@pdv/errors";
import {
  ProductResponseSchema,
  UpsertProductInputSchema,
} from "../schemas/product.schema";
import type {
  ProductResponse,
  UpsertProductInput,
} from "../interfaces/product.interface";
import { productKeys } from "../queries/product.query";

function createProduct(data: UpsertProductInput): Promise<ProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: "/products",
    requestSchema: UpsertProductInputSchema,
    responseSchema: ProductResponseSchema,
  });

  return api(data).catch(mapApiError);
}

function updateProduct(
  id: string,
  data: UpsertProductInput,
): Promise<ProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.PUT,
    path: `/products/${id}`,
    requestSchema: UpsertProductInputSchema,
    responseSchema: ProductResponseSchema,
  });

  return api(data).catch(mapApiError);
}

function activateProduct(id: string): Promise<ProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: `/products/${id}/activate`,
    requestSchema: z.object({}),
    responseSchema: ProductResponseSchema,
  });

  return api({}).catch(mapApiError);
}

function deactivateProduct(id: string): Promise<ProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.POST,
    path: `/products/${id}/deactivate`,
    requestSchema: z.object({}),
    responseSchema: ProductResponseSchema,
  });

  return api({}).catch(mapApiError);
}

export function useCreateProductMutation(): UseMutationResult<
  ProductResponse,
  Error,
  UpsertProductInput
> {
  const queryClient: QueryClient = useQueryClient();

  return useMutation({
    mutationFn: createProduct,
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: productKeys.lists() });
    },
  });
}

export function useUpdateProductMutation(): UseMutationResult<
  ProductResponse,
  Error,
  { id: string; data: UpsertProductInput }
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => updateProduct(id, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: productKeys.detail(result.id),
      });
      queryClient.invalidateQueries({ queryKey: productKeys.lists() });
    },
  });
}

export function useActivateProductMutation(): UseMutationResult<
  ProductResponse,
  Error,
  string
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: activateProduct,
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: productKeys.detail(result.id),
      });
      queryClient.invalidateQueries({ queryKey: productKeys.lists() });
    },
  });
}

export function useDeactivateProductMutation(): UseMutationResult<
  ProductResponse,
  Error,
  string
> {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deactivateProduct,
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: productKeys.detail(result.id),
      });
      queryClient.invalidateQueries({ queryKey: productKeys.lists() });
    },
  });
}
