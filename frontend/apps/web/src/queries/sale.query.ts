import { createApiCall, HttpMethod } from "@pdv/http";
import { z } from "zod";
import { mapApiError } from "@pdv/errors";
import { useQuery } from "@tanstack/react-query";
import type { UseQueryResult } from "@tanstack/react-query";
import {
  SaleResponseSchema,
  SaleListResponseSchema,
  ListSalesParamsSchema,
} from "../schemas/sale.schema";
import type {
  SaleResponse,
  SaleListResponse,
  ListSalesParams,
} from "../interfaces/sale.interface";

function listSales(params: ListSalesParams = {}): Promise<SaleListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/sales",
    requestSchema: ListSalesParamsSchema,
    responseSchema: SaleListResponseSchema,
  });

  return api(params).catch(mapApiError);
}

function getSale(id: string): Promise<SaleResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/sales/${id}`,
    requestSchema: z.object({}),
    responseSchema: SaleResponseSchema,
  });

  return api({}).catch(mapApiError);
}

export const saleKeys = {
  all: ["sales"] as const,
  lists: () => [...saleKeys.all, "list"] as const,
  list: (params?: ListSalesParams) => [...saleKeys.lists(), params] as const,
  details: () => [...saleKeys.all, "detail"] as const,
  detail: (id: string) => [...saleKeys.details(), id] as const,
};

export function useListSalesQuery(
  params?: ListSalesParams,
): UseQueryResult<SaleListResponse> {
  return useQuery({
    queryKey: saleKeys.list(params),
    queryFn: () => listSales(params),
  });
}

export function useGetSaleQuery(id: string): UseQueryResult<SaleResponse> {
  return useQuery({
    queryKey: saleKeys.detail(id),
    queryFn: () => getSale(id),
    enabled: !!id,
  });
}
