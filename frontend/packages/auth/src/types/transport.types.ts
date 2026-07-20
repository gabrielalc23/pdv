export interface ApiErrorLike {
  status: number | null;
  code: string | null;
}

export interface AuthTransportConfiguration {
  getAccessToken: () => string | null;
  getCsrfToken: () => string | null;
  refresh: () => Promise<void>;
  shouldRefresh: (error: ApiErrorLike) => boolean;
  shouldInvalidateAuth: (error: ApiErrorLike) => boolean;
  onAuthLost: (error: ApiErrorLike) => void | Promise<void>;
}

export interface AuthRetryRequestConfig {
  _authRetry?: boolean;
}

export type AuthLockCallback = () => Promise<void>;

export type AuthBroadcastMessage =
  | {
      type: "token-updated";
      sourceId: string;
      accessToken: string;
      expiresIn: number;
      session: unknown;
    }
  | {
      type: "logout";
      sourceId: string;
    }
  | {
      type: "auth-lost";
      sourceId: string;
      reason: string;
    }
  | {
      type: "context-changed";
      sourceId: string;
      accessToken: string;
      expiresIn: number;
      session: unknown;
    };
