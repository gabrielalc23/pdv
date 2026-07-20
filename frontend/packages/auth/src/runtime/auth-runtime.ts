import type { AxiosInstance } from "axios";
import type { AuthSessionResponse } from "../schemas/auth-response.schema";
import type { AuthBootstrapResult, AuthLossReason } from "../types/auth.types";
import type {
  AuthTransportConfiguration,
  ApiErrorLike,
} from "../types/transport.types";
import type { Clock } from "../stores/access-token.store";

import { getAccessToken, getCsrfToken } from "../stores";
import { clearAccessToken } from "../stores/access-token.store";
import {
  getAuthSessionState,
  resetAuthSession,
} from "../stores/auth-session.store";
import { clearCsrfToken } from "../stores/csrf-token.store";
import { resetAuthEventBus } from "../events/auth-events";
import {
  configureAuthTransport,
  isRefreshEligibleError,
  isTerminalAuthError,
  isExpectedAnonymousRefreshError,
} from "./configure-http";
import { bootstrapAuth } from "./bootstrap";
import { fetchCsrfToken } from "../client/csrf.client";
import { refreshSession } from "../client/refresh.client";
import { clearLocalAuthState } from "../client/logout.client";
import { applyContextChange } from "../client/context.client";
import type {
  BroadcastAdapter,
  LockAdapter,
} from "../coordinator/coordinator.types";
import { createBroadcastAdapter } from "../coordinator/broadcast-adapter";
import { createLockAdapter } from "../coordinator/lock-adapter";
import type { RefreshCoordinator } from "../coordinator/refresh-coordinator";
import { createRefreshCoordinator } from "../coordinator/refresh-coordinator";

let globalGeneration = 0;

export interface AuthRuntimeOptions {
  publicInstance: AxiosInstance;
  channelName?: string;
  refreshLockName?: string;
  clock?: Clock;
  broadcast?: BroadcastAdapter;
  lockManager?: LockAdapter;
  sourceId?: string;
  onAuthLost?: (reason: AuthLossReason) => void | Promise<void>;
}

export interface AuthRuntime {
  bootstrap: () => Promise<AuthBootstrapResult>;
  refresh: () => Promise<AuthSessionResponse>;
  fetchCsrf: () => Promise<string>;
  clear: (reason?: AuthLossReason) => void;
  publishLogout: VoidFunction;
  publishContextChange: (session: AuthSessionResponse) => void;
  destroy: VoidFunction;
}

export function createAuthRuntime(options: AuthRuntimeOptions): AuthRuntime {
  const {
    publicInstance,
    channelName = "pdv-auth",
    sourceId = typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`,
    onAuthLost,
    broadcast: broadcastOverride,
    lockManager: lockOverride,
  } = options;

  const broadcast: BroadcastAdapter =
    broadcastOverride ?? createBroadcastAdapter(channelName);
  const lock: LockAdapter = lockOverride ?? createLockAdapter();
  const generationRef = { current: 0 };

  let isDestroyed = false;

  const coordinator: RefreshCoordinator = createRefreshCoordinator({
    publicInstance,
    broadcast,
    lock,
    sourceId,
    getGeneration: () => generationRef.current,
  });

  async function refreshAndGetSession(): Promise<AuthSessionResponse> {
    return refreshSession(publicInstance);
  }

  async function fetchCsrfAndStore(): Promise<string> {
    return fetchCsrfToken(publicInstance);
  }

  const transportConfig: AuthTransportConfiguration = {
    getAccessToken: () => {
      if (isDestroyed) return null;
      return getAccessToken();
    },
    getCsrfToken: () => {
      if (isDestroyed) return null;
      return getCsrfToken();
    },
    refresh: async () => {
      if (isDestroyed) throw new Error("Runtime destroyed");
      await coordinator.refresh();
      generationRef.current = ++globalGeneration;
    },
    shouldRefresh: (error: ApiErrorLike): boolean => {
      if (isDestroyed) return false;
      return isRefreshEligibleError(error);
    },
    shouldInvalidateAuth: (error: ApiErrorLike): boolean => {
      if (isDestroyed) return false;
      if (isTerminalAuthError(error)) return true;
      return isExpectedAnonymousRefreshError(error);
    },
    onAuthLost: async (error: ApiErrorLike): Promise<void> => {
      if (isDestroyed) return;
      const reason: AuthLossReason = mapCodeToReason(error.code);
      clearLocalAuthState(reason);
      coordinator.publishAuthLost(reason);
      if (onAuthLost) {
        await onAuthLost(reason);
      }
    },
  };

  const cleanupHttp = configureAuthTransport(transportConfig);

  const runtime: AuthRuntime = {
    async bootstrap(): Promise<AuthBootstrapResult> {
      if (isDestroyed) throw new Error("Runtime destroyed");
      return bootstrapAuth({
        publicInstance,
        doRefresh: refreshAndGetSession,
        fetchCsrf: fetchCsrfAndStore,
        sourceId,
      });
    },

    async refresh(): Promise<AuthSessionResponse> {
      if (isDestroyed) throw new Error("Runtime destroyed");
      await coordinator.refresh();
      const sessionState = getAuthSessionState().session;
      if (!sessionState)
        throw new Error("Refresh completed but no session available");
      return sessionState;
    },

    async fetchCsrf(): Promise<string> {
      if (isDestroyed) throw new Error("Runtime destroyed");
      return fetchCsrfAndStore();
    },

    clear(reason: AuthLossReason = "logout"): void {
      clearLocalAuthState(reason);
      coordinator.publishAuthLost(reason);
    },

    publishLogout(): void {
      coordinator.publishLogout();
    },

    publishContextChange(session: AuthSessionResponse): void {
      const applied = applyContextChange(session);
      coordinator.publishContextChange(applied);
    },

    destroy(): void {
      if (isDestroyed) return;
      isDestroyed = true;
      cleanupHttp();
      coordinator.destroy();
      broadcast.close();
      clearAccessToken();
      resetAuthSession();
      clearCsrfToken();
      resetAuthEventBus();
    },
  };

  return runtime;
}

function mapCodeToReason(code: string | null): AuthLossReason {
  switch (code) {
    case "SESSION_REVOKED":
      return "session_revoked";
    case "SESSION_EXPIRED":
      return "session_expired";
    case "REFRESH_TOKEN_REUSED":
      return "refresh_token_reused";
    case "REFRESH_TOKEN_INVALID":
    case "REFRESH_TOKEN_EXPIRED":
      return "refresh_failed";
    default:
      return "unknown";
  }
}
