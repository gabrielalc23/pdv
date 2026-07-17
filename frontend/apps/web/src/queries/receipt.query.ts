import { createApiCall, HttpMethod } from "@pdv/http";
import { z } from "zod";
import { mapApiError } from "@pdv/errors";
import { useQuery } from "@tanstack/react-query";
import type { UseQueryResult } from "@tanstack/react-query";
import { ReceiptResponseSchema } from "../schemas/receipt.schema";
import type { ReceiptResponse } from "../interfaces/receipt.interface";

function getReceipt(saleId: string): Promise<ReceiptResponse> {
  const api = createApiCall({
    type: "public",
    method: HttpMethod.GET,
    path: `/sales/${saleId}/receipt`,
    requestSchema: z.object({}),
    responseSchema: ReceiptResponseSchema,
  });

  return api({}).catch(mapApiError);
}

export const receiptKeys = {
  all: ["receipt"] as const,
  detail: (saleId: string) => [...receiptKeys.all, saleId] as const,
};

export function useGetReceiptQuery(
  saleId: string,
): UseQueryResult<ReceiptResponse> {
  return useQuery({
    queryKey: receiptKeys.detail(saleId),
    queryFn: () => getReceipt(saleId),
    enabled: !!saleId,
  });
}
