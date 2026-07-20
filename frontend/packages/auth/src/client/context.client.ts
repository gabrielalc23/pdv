import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import { AuthSessionResponseSchema } from "../schemas/auth-response.schema";
import { setAccessToken } from "../stores/access-token.store";
import { setAuthenticatedSession } from "../stores/auth-session.store";
import { dispatchAuthEvent } from "../events/auth-events";

export function applyContextChange(raw: unknown): AuthSessionResponse {
  const parsed: AuthSessionResponse = AuthSessionResponseSchema.parse(raw);
  setAccessToken(parsed.accessToken, parsed.expiresIn);
  setAuthenticatedSession(parsed);
  dispatchAuthEvent({ type: "token-updated" });
  dispatchAuthEvent({ type: "context-changed", session: parsed });
  return parsed;
}
