import axios from "axios";
import type { AxiosInstance, InternalAxiosRequestConfig } from "axios";

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

instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig<unknown>): InternalAxiosRequestConfig<unknown> => {
    const token: string | null = localStorage.getItem("accessToken");

    if (token) {
      config.headers.set("Authorization", `Bearer ${token}`);
    }

    return config;
  },
  (error) => Promise.reject(error),
);
