import axios from "axios"
import type { AxiosInstance, InternalAxiosRequestConfig } from "axios"

const API_BASE_URL: string = process.env.NEXT_PUBLIC_API_URL || "https://api.seusite.com"

export const instance: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
})

instance.interceptors.request.use(
  (config: InternalAxiosRequestConfig<unknown>): InternalAxiosRequestConfig<unknown> => {
    const token: string | null = localStorage.getItem("accessToken")

    if (token) {
      config.headers.set("Authorization", `Bearer ${token}`)
    }

    return config
  },
  (error) => Promise.reject(error),
)
