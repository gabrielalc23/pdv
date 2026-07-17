import { createApiCall, HttpMethod } from "@pdv/http";
import { z } from "zod";
import { mapApiError } from "@pdv/errors";
import { useQuery } from "@tanstack/react-query";
import type { UseQueryResult } from "@tanstack/react-query";
import { FiscalDocumentResponseSchema } from "../schemas/fiscal.schema";
import type { FiscalDocumentResponse } from "../interfaces/fiscal.interface";

function getFiscalDocument(saleId: string): Promise<FiscalDocumentResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/sales/${saleId}/fiscal-document`,
    requestSchema: z.object({}),
    responseSchema: FiscalDocumentResponseSchema,
  });

  return api({}).catch(mapApiError);
}

export const fiscalKeys = {
  all: ["fiscal"] as const,
  document: (saleId: string) => [...fiscalKeys.all, saleId] as const,
};

export function useGetFiscalDocumentQuery(
  saleId: string,
): UseQueryResult<FiscalDocumentResponse> {
  return useQuery({
    queryKey: fiscalKeys.document(saleId),
    queryFn: () => getFiscalDocument(saleId),
    enabled: !!saleId,
  });
}
