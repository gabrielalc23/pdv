import type { AuthSessionResponse } from "../schemas/auth-response.schema";

export type AuthBootstrapResult =
  | {
      status: "authenticated";
      session: AuthSessionResponse;
    }
  | {
      status: "anonymous";
    }
  | {
      status: "unavailable";
      error: Error;
    };

export type AuthLossReason =
  | "session_revoked"
  | "session_expired"
  | "refresh_token_reused"
  | "refresh_failed"
  | "logout"
  | "unknown";

export type AuthStatus = "unknown" | "anonymous" | "authenticated";
