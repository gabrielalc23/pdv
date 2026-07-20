import axios from "axios";
import type {
  AxiosInstance,
  InternalAxiosRequestConfig,
  AxiosResponse,
  AxiosError,
} from "axios";
import { getAuthTransportConfiguration } from "./auth-transport-configuration";
import type {
  AuthRetryRequestConfig,
  ApiErrorLike,
} from "./auth-transport-configuration";

const API_BASE_URL: string =
  import.meta.env.VITE_API_URL ??
  (import.meta.env.MODE === "test" ? "http://localhost:3000" : "/api");

export const instance: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

const UNSAFE_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);

// Request interceptor: attach Bearer token and CSRF token
instance.interceptors.request.use(
  (
    config: InternalAxiosRequestConfig<unknown> & AuthRetryRequestConfig,
  ): InternalAxiosRequestConfig<unknown> => {
    const transportConfig = getAuthTransportConfiguration();

    // Attach Bearer token
    if (transportConfig) {
      const token: string | null = transportConfig.getAccessToken();
      if (token && typeof config.headers !== "undefined") {
        config.headers.set("Authorization", `Bearer ${token}`);
      }
    }

    // Attach CSRF token for unsafe methods
    if (
      transportConfig &&
      config.method &&
      UNSAFE_METHODS.has(config.method.toUpperCase())
    ) {
      const csrfToken: string | null = transportConfig.getCsrfToken();
      if (csrfToken) {
        config.headers.set("X-CSRF-Token", csrfToken);
      }
    }

    return config;
  },
  (error) => Promise.reject(error),
);

// Response interceptor: handle auth errors, refresh, retry
instance.interceptors.response.use(
  (response: AxiosResponse<unknown>): AxiosResponse<unknown> => {
    return response;
  },
  async (
    error: AxiosError<{ error?: { code?: string; message?: string } }>,
  ): Promise<unknown> => {
    const transportConfig = getAuthTransportConfiguration();

    if (!transportConfig) {
      return Promise.reject(error);
    }

    const originalRequest = error.config as
      | (InternalAxiosRequestConfig<unknown> & AuthRetryRequestConfig)
      | undefined;

    if (!originalRequest) {
      return Promise.reject(error);
    }

    const errResponse = error.response;
    const status: number | null = errResponse?.status ?? null;
    const errData = errResponse?.data;
    const code: string | null = errData?.error?.code ?? null;

    const apiError: ApiErrorLike = { status, code };

    // Do not intercept /auth/refresh itself
    const url: string | undefined = originalRequest.url;
    if (url && (url === "/auth/refresh" || url.endsWith("/auth/refresh"))) {
      return Promise.reject(error);
    }

    // Already retried
    if (originalRequest._authRetry) {
      return Promise.reject(error);
    }

    // Check if this is a refresh-eligible error
    if (transportConfig.shouldRefresh(apiError)) {
      try {
        await transportConfig.refresh();
      } catch (refreshError: unknown) {
        const refreshErr = refreshError as AxiosError<{
          error?: { code?: string };
        }>;
        const refreshResponse = refreshErr.response;
        const refreshData = refreshResponse?.data;
        const refreshCode: string | null = refreshData?.error?.code ?? null;
        const refreshStatus: number | null = refreshResponse?.status ?? null;
        if (
          transportConfig.shouldInvalidateAuth({
            status: refreshStatus,
            code: refreshCode,
          })
        ) {
          await transportConfig.onAuthLost({
            status: refreshStatus,
            code: refreshCode,
          });
        }
        return Promise.reject(error);
      }

      // Retry once with new token
      const newToken: string | null = transportConfig.getAccessToken();
      if (newToken) {
        originalRequest.headers.set("Authorization", `Bearer ${newToken}`);
      }

      const csrfToken: string | null = transportConfig.getCsrfToken();
      if (
        csrfToken &&
        originalRequest.method &&
        UNSAFE_METHODS.has(originalRequest.method.toUpperCase())
      ) {
        originalRequest.headers.set("X-CSRF-Token", csrfToken);
      }

      originalRequest._authRetry = true;

      return instance.request(originalRequest);
    }

    // Check if this is a terminal auth error (should invalidate)
    if (transportConfig.shouldInvalidateAuth(apiError)) {
      await transportConfig.onAuthLost(apiError);
      return Promise.reject(error);
    }

    return Promise.reject(error);
  },
);
