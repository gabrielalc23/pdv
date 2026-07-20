import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import type { AuthRetryRequestConfig } from "@pdv/http";
import { getCsrfToken, clearCsrfToken } from "../stores/csrf-token.store";
import { clearAccessToken } from "../stores/access-token.store";
import { resetAuthSession } from "../stores/auth-session.store";
import { dispatchAuthEvent } from "../events/auth-events";
import type { AuthLossReason } from "../types/auth.types";

export async function logout(publicInstance: AxiosInstance): Promise<void> {
  const csrfToken: string | null = getCsrfToken();
  const headers: Record<string, string> = {};

  if (csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  try {
    const config: AxiosRequestConfig & AuthRetryRequestConfig = {
      method: "POST",
      url: "/auth/logout",
      headers,
      withCredentials: true,
      _authRetry: true,
    };
    const response: AxiosResponse<unknown> =
      await publicInstance.request(config);
    if (response.status !== 204) {
      // still ok, proceed with cleanup
    }
  } catch {
    // Even on network failure, clear local state
  }

  clearLocalAuthState("logout");
}

export function clearLocalAuthState(reason: AuthLossReason = "logout"): void {
  clearAccessToken();
  resetAuthSession();
  clearCsrfToken();
  dispatchAuthEvent({ type: "auth-lost", reason });
  dispatchAuthEvent({ type: "logout" });
}
