import type { AxiosInstance, AxiosResponse } from "axios";
import { AuthSessionResponseSchema } from "../schemas/auth-response.schema";
import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import { setAccessToken } from "../stores/access-token.store";
import { setAuthenticatedSession } from "../stores/auth-session.store";
import { getCsrfToken } from "../stores/csrf-token.store";
import { dispatchAuthEvent } from "../events/auth-events";

export async function refreshSession(
  publicInstance: AxiosInstance,
): Promise<AuthSessionResponse> {
  const csrfToken: string | null = getCsrfToken();
  const headers: Record<string, string> = {};

  if (csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  const response: AxiosResponse<unknown> = await publicInstance.request({
    method: "POST",
    url: "/auth/refresh",
    headers,
    withCredentials: true,
  });

  const parsed: AuthSessionResponse = AuthSessionResponseSchema.parse(
    response.data,
  );

  setAccessToken(parsed.accessToken, parsed.expiresIn);
  setAuthenticatedSession(parsed);
  dispatchAuthEvent({ type: "token-updated" });
  dispatchAuthEvent({ type: "session-updated", session: parsed });

  return parsed;
}
