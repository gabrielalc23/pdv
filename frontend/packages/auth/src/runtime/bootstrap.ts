import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import type { AxiosInstance } from "axios";
import type { AuthBootstrapResult } from "../types/auth.types";
import { setAccessToken } from "../stores/access-token.store";
import {
  setAuthenticatedSession,
  setAnonymousSession,
  resetAuthSession,
} from "../stores/auth-session.store";
import { dispatchAuthEvent } from "../events/auth-events";
import { isExpectedAnonymousRefreshError } from "./configure-http";

export interface BootstrapOptions {
  publicInstance: AxiosInstance;
  doRefresh: () => Promise<AuthSessionResponse>;
  fetchCsrf: () => Promise<string>;
  sourceId: string;
}

export async function bootstrapAuth(
  options: BootstrapOptions,
): Promise<AuthBootstrapResult> {
  try {
    // Step 1: Fetch CSRF token
    try {
      await options.fetchCsrf();
    } catch {
      // CSRF fetch failure is not fatal for bootstrap
    }

    // Step 2: Try to refresh (get a session from the refresh cookie)
    try {
      const session: AuthSessionResponse = await options.doRefresh();
      setAccessToken(session.accessToken, session.expiresIn);
      setAuthenticatedSession(session);
      dispatchAuthEvent({ type: "bootstrap-completed", authenticated: true });
      return { status: "authenticated", session };
    } catch (err: unknown) {
      const error = err as {
        response?: { status?: number; data?: { error?: { code?: string } } };
      };
      const code: string | undefined = error.response?.data?.error?.code;

      // REFRESH_TOKEN_MISSING means no session - this is expected anonymous
      if (code && isExpectedAnonymousRefreshError({ status: 401, code })) {
        resetAuthSession();
        setAnonymousSession();
        dispatchAuthEvent({
          type: "bootstrap-completed",
          authenticated: false,
        });
        return { status: "anonymous" };
      }

      // session expired/revoked means anonymous
      if (code && ["SESSION_EXPIRED", "SESSION_REVOKED"].includes(code)) {
        resetAuthSession();
        setAnonymousSession();
        dispatchAuthEvent({
          type: "bootstrap-completed",
          authenticated: false,
        });
        return { status: "anonymous" };
      }

      // network or other error - return unavailable
      const bootstrapError =
        err instanceof Error ? err : new Error("Bootstrap failed");
      dispatchAuthEvent({ type: "bootstrap-completed", authenticated: false });
      return { status: "unavailable", error: bootstrapError };
    }
  } catch (err: unknown) {
    const bootstrapError =
      err instanceof Error ? err : new Error("Bootstrap failed unexpectedly");
    dispatchAuthEvent({ type: "bootstrap-completed", authenticated: false });
    return { status: "unavailable", error: bootstrapError };
  }
}
