import { createApiCall, HttpMethod } from "@pdv/http"
import { z } from "zod"
import { mapApiError } from "@pdv/errors"
import { useQuery } from "@tanstack/react-query"
import type { UseQueryResult } from "@tanstack/react-query"
import { HealthResponseSchema } from "../schemas/health.schema"
import type { HealthResponse } from "../interfaces/health.interface"

function checkHealth(): Promise<HealthResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: "/health",
    requestSchema: z.object({}),
    responseSchema: HealthResponseSchema,
  })

  return api({}).catch(mapApiError)
}

export const healthKeys = {
  all: ["health"] as const,
}

export function useHealthQuery(): UseQueryResult<HealthResponse> {
  return useQuery({
    queryKey: healthKeys.all,
    queryFn: checkHealth,
  })
}
