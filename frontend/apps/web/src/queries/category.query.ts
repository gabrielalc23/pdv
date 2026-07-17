import { createApiCall, HttpMethod } from "@pdv/http";
import { mapApiError } from "@pdv/errors";
import { z } from "zod";
import { useQuery } from "@tanstack/react-query";
import type { UseQueryResult } from "@tanstack/react-query";
import { CategoryListResponseSchema } from "../schemas/category.schema";
import type {
  CategoryListResponse,
  ListCategoriesParams,
} from "../interfaces/category.interface";

function listCategories(
  params: ListCategoriesParams = {},
): Promise<CategoryListResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/categories",
    requestSchema: z.object({
      search: z.string().optional(),
      activeOnly: z.boolean().optional(),
    }),
    responseSchema: CategoryListResponseSchema,
  });

  return api(params).catch(mapApiError);
}

export const categoryKeys = {
  all: ["categories"] as const,
  lists: () => [...categoryKeys.all, "list"] as const,
  list: (params?: ListCategoriesParams) =>
    [...categoryKeys.lists(), params] as const,
};

export function useListCategoriesQuery(
  params?: ListCategoriesParams,
): UseQueryResult<CategoryListResponse> {
  return useQuery({
    queryKey: categoryKeys.list(params),
    queryFn: () => listCategories(params),
  });
}
