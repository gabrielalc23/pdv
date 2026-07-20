import type { AxiosInstance, AxiosResponse } from "axios";
import { CsrfResponseSchema } from "../schemas/csrf.schema";
import { setCsrfToken } from "../stores/csrf-token.store";

export async function fetchCsrfToken(
  publicInstance: AxiosInstance,
): Promise<string> {
  const response: AxiosResponse<unknown> = await publicInstance.request({
    method: "GET",
    url: "/auth/csrf",
    withCredentials: true,
  });

  const parsed = CsrfResponseSchema.parse(response.data);
  setCsrfToken(parsed.csrfToken);
  return parsed.csrfToken;
}
