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

let currentConfig: AuthTransportConfiguration | null = null;

export function configureAuthTransport(
  config: AuthTransportConfiguration,
): () => void {
  currentConfig = config;
  return (): void => {
    if (currentConfig === config) {
      currentConfig = null;
    }
  };
}

export function resetAuthTransportConfiguration(): void {
  currentConfig = null;
}

export function getAuthTransportConfiguration(): AuthTransportConfiguration | null {
  return currentConfig;
}
