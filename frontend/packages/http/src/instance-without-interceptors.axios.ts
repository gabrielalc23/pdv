import axios from "axios";
import type { AxiosInstance, InternalAxiosRequestConfig } from "axios";
import { getAuthTransportConfiguration } from "./auth-transport-configuration";
import type { AuthRetryRequestConfig } from "./auth-transport-configuration";

const API_BASE_URL: string =
  import.meta.env.VITE_API_URL ??
  (import.meta.env.MODE === "test" ? "http://localhost:3000" : "/api");

const UNSAFE_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);

export const instanceWithoutInterceptors: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

// Only attach CSRF token for unsafe methods; no Bearer token
instanceWithoutInterceptors.interceptors.request.use(
  (
    config: InternalAxiosRequestConfig<unknown> & AuthRetryRequestConfig,
  ): InternalAxiosRequestConfig<unknown> => {
    const transportConfig = getAuthTransportConfiguration();

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
