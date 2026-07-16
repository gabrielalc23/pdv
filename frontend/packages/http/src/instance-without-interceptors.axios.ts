import axios from "axios"
import type { AxiosInstance } from "axios"

const API_BASE_URL: string = process.env.NEXT_PUBLIC_API_URL || "https://api.seusite.com"

export const instanceWithoutInterceptors: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
})
