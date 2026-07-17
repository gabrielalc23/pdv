import axios from "axios";
import type { AxiosInstance } from "axios";

const API_BASE_URL: string =
  import.meta.env.VITE_API_URL ??
  (import.meta.env.MODE === "test" ? "http://localhost:3000" : "/api");

export const instanceWithoutInterceptors: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});
