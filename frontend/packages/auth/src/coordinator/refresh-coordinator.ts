import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import { AuthSessionResponseSchema } from "../schemas/auth-response.schema";
import { setAccessToken, clearAccessToken } from "../stores/access-token.store";
import {
  setAuthenticatedSession,
  setAnonymousSession,
} from "../stores/auth-session.store";
import { clearLocalAuthState } from "../client/logout.client";
import { dispatchAuthEvent } from "../events/auth-events";
import { AuthBroadcastMessageSchema } from "../events/auth-message.schema";
import type { AuthBroadcastMessageParsed } from "../events/auth-message.schema";
import { refreshSession } from "../client/refresh.client";
import type { BroadcastAdapter, LockAdapter } from "./coordinator.types";

let generation = 0;

function nextGeneration(): number {
  return ++generation;
}

export interface RefreshCoordinatorOptions {
  publicInstance: Parameters<typeof refreshSession>[0];
  broadcast: BroadcastAdapter;
  lock: LockAdapter;
  channelName?: string;
  sourceId: string;
  getGeneration: () => number;
}

export interface RefreshCoordinator {
  refresh: () => Promise<void>;
  publishToken: (session: AuthSessionResponse) => void;
  publishLogout: VoidFunction;
  publishAuthLost: (reason: string) => void;
  publishContextChange: (session: AuthSessionResponse) => void;
  destroy: VoidFunction;
}

export function createRefreshCoordinator(
  options: RefreshCoordinatorOptions,
): RefreshCoordinator {
  const { publicInstance, broadcast, lock, sourceId } = options;

  let localGen = 0;
  let refreshInFlight: Promise<void> | null = null;
  let isDestroyed = false;
  const cleanupFns: Array<() => void> = [];

  const cleanupMessages = broadcast.onMessage((raw: unknown) => {
    if (isDestroyed) return;

    const parsedResult = AuthBroadcastMessageSchema.safeParse(raw);
    if (!parsedResult.success) return;

    const msg: AuthBroadcastMessageParsed = parsedResult.data;

    if (msg.sourceId === sourceId) return;

    switch (msg.type) {
      case "token-updated": {
        try {
          const session = AuthSessionResponseSchema.parse(msg.session);
          setAccessToken(msg.accessToken, msg.expiresIn);
          setAuthenticatedSession(session);
          localGen = nextGeneration();
          dispatchAuthEvent({ type: "token-updated" });
          dispatchAuthEvent({ type: "session-updated", session });
        } catch {
          // ignore invalid session from broadcast
        }
        break;
      }
      case "logout": {
        clearLocalAuthState("logout");
        break;
      }
      case "auth-lost": {
        clearLocalAuthState("unknown");
        break;
      }
      case "context-changed": {
        try {
          const session = AuthSessionResponseSchema.parse(msg.session);
          setAccessToken(msg.accessToken, msg.expiresIn);
          setAuthenticatedSession(session);
          localGen = nextGeneration();
          dispatchAuthEvent({ type: "token-updated" });
          dispatchAuthEvent({ type: "context-changed", session });
        } catch {
          // ignore invalid session
        }
        break;
      }
    }
  });
  cleanupFns.push(cleanupMessages);

  async function doRefresh(): Promise<void> {
    await lock.acquire("pdv-auth-refresh", async () => {
      const currentGen = options.getGeneration();
      if (currentGen > localGen) {
        return;
      }

      let session: AuthSessionResponse;
      try {
        session = await refreshSession(publicInstance);
      } catch (err: unknown) {
        const error = err as {
          response?: { status?: number; data?: { error?: { code?: string } } };
        };
        const code: string | undefined = error.response?.data?.error?.code;
        if (
          code &&
          [
            "SESSION_REVOKED",
            "SESSION_EXPIRED",
            "REFRESH_TOKEN_REUSED",
            "REFRESH_TOKEN_INVALID",
            "REFRESH_TOKEN_EXPIRED",
          ].includes(code)
        ) {
          clearLocalAuthState("session_revoked");
          broadcast.postMessage({
            type: "auth-lost",
            sourceId,
            reason: "session_revoked",
          });
        } else if (code === "REFRESH_TOKEN_MISSING") {
          clearAccessToken();
          setAnonymousSession();
          dispatchAuthEvent({ type: "auth-lost", reason: "session_expired" });
        }
        throw err;
      }

      setAccessToken(session.accessToken, session.expiresIn);
      setAuthenticatedSession(session);
      localGen = nextGeneration();

      broadcast.postMessage({
        type: "token-updated",
        sourceId,
        accessToken: session.accessToken,
        expiresIn: session.expiresIn,
        session,
      });
    });
  }

  const coordinator: RefreshCoordinator = {
    async refresh(): Promise<void> {
      if (isDestroyed) throw new Error("RefreshCoordinator is destroyed");

      if (refreshInFlight) {
        return refreshInFlight;
      }

      refreshInFlight = doRefresh()
        .then(() => {
          refreshInFlight = null;
        })
        .catch((err: unknown) => {
          refreshInFlight = null;
          throw err;
        });

      return refreshInFlight;
    },

    publishToken(session: AuthSessionResponse): void {
      broadcast.postMessage({
        type: "token-updated",
        sourceId,
        accessToken: session.accessToken,
        expiresIn: session.expiresIn,
        session,
      });
    },

    publishLogout(): void {
      broadcast.postMessage({ type: "logout", sourceId });
    },

    publishAuthLost(reason: string): void {
      broadcast.postMessage({ type: "auth-lost", sourceId, reason });
    },

    publishContextChange(session: AuthSessionResponse): void {
      broadcast.postMessage({
        type: "context-changed",
        sourceId,
        accessToken: session.accessToken,
        expiresIn: session.expiresIn,
        session,
      });
    },

    destroy(): void {
      isDestroyed = true;
      refreshInFlight = null;
      for (const fn of cleanupFns) {
        fn();
      }
      cleanupFns.length = 0;
    },
  };

  return coordinator;
}
