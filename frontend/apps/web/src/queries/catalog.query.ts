import { createApiCall, HttpMethod } from "@pdv/http";
import { z } from "zod";
import { mapApiError } from "@pdv/errors";
import { useQuery } from "@tanstack/react-query";
import type { UseQueryResult } from "@tanstack/react-query";
import {
  CatalogProductResponseSchema,
  CatalogListResponseSchema,
  ListCatalogParamsSchema,
} from "../schemas/catalog.schema";
import type {
  CatalogProductResponse,
  CatalogListResponse,
  ListCatalogParams,
} from "../interfaces/catalog.interface";

function listCatalog(
  params: ListCatalogParams = {},
): Promise<CatalogListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/catalog",
    requestSchema: ListCatalogParamsSchema,
    responseSchema: CatalogListResponseSchema,
  });

  return api(params).catch(mapApiError);
}

function getCatalogProduct(id: string): Promise<CatalogProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/catalog/${id}`,
    requestSchema: z.object({}),
    responseSchema: CatalogProductResponseSchema,
  });

  return api({}).catch(mapApiError);
}

function getCatalogProductByBarcode(
  barcode: string,
): Promise<CatalogProductResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/catalog/barcode/${barcode}`,
    requestSchema: z.object({}),
    responseSchema: CatalogProductResponseSchema,
  });

  return api({}).catch(mapApiError);
}

export const catalogKeys = {
  all: ["catalog"] as const,
  lists: () => [...catalogKeys.all, "list"] as const,
  list: (params?: ListCatalogParams) =>
    [...catalogKeys.lists(), params] as const,
  details: () => [...catalogKeys.all, "detail"] as const,
  detail: (id: string) => [...catalogKeys.details(), id] as const,
  barcode: (barcode: string) =>
    [...catalogKeys.all, "barcode", barcode] as const,
};

export function useListCatalogQuery(
  params?: ListCatalogParams,
): UseQueryResult<CatalogListResponse> {
  return useQuery({
    queryKey: catalogKeys.list(params),
    queryFn: () => listCatalog(params),
  });
}

export function useGetCatalogProductQuery(
  id: string,
): UseQueryResult<CatalogProductResponse> {
  return useQuery({
    queryKey: catalogKeys.detail(id),
    queryFn: () => getCatalogProduct(id),
    enabled: !!id,
  });
}

export function useGetCatalogProductByBarcodeQuery(
  barcode: string,
): UseQueryResult<CatalogProductResponse> {
  return useQuery({
    queryKey: catalogKeys.barcode(barcode),
    queryFn: () => getCatalogProductByBarcode(barcode),
    enabled: !!barcode,
  });
}
