import { createApiCall, HttpMethod } from "@pdv/http"
import { z } from "zod"
import { mapApiError } from "@pdv/errors"
import { useQuery } from "@tanstack/react-query"
import type { UseQueryResult } from "@tanstack/react-query"
import {
  ProductResponseSchema,
  ProductListResponseSchema,
  ListProductsParamsSchema,
} from "../schemas/product.schema"
import type {
  ProductResponse,
  ProductListResponse,
  ListProductsParams,
} from "../interfaces/product.interface"

function listProducts(params: ListProductsParams = {}): Promise<ProductListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/products",
    requestSchema: ListProductsParamsSchema,
    responseSchema: ProductListResponseSchema,
  })

  return api(params).catch(mapApiError)
}

function getProduct(id: string): Promise<ProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/products/${id}`,
    requestSchema: z.object({}),
    responseSchema: ProductResponseSchema,
  })

  return api({}).catch(mapApiError)
}

export const productKeys = {
  all: ["products"] as const,
  lists: () => [...productKeys.all, "list"] as const,
  list: (params?: ListProductsParams) => [...productKeys.lists(), params] as const,
  details: () => [...productKeys.all, "detail"] as const,
  detail: (id: string) => [...productKeys.details(), id] as const,
}

export function useListProductsQuery(
  params?: ListProductsParams,
): UseQueryResult<ProductListResponse> {
  return useQuery({
    queryKey: productKeys.list(params),
    queryFn: () => listProducts(params),
  })
}

export function useGetProductQuery(id: string): UseQueryResult<ProductResponse> {
  return useQuery({
    queryKey: productKeys.detail(id),
    queryFn: () => getProduct(id),
    enabled: !!id,
  })
}
