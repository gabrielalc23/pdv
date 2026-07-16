import { isAxiosError } from "axios"
import type { AxiosError } from "axios"
import { AppError } from "./app.error"
import { ApiError } from "./api.error"
import { NotFoundError } from "./not-found.error"
import { ConflictError } from "./conflict.error"
import { ValidationError } from "./validation.error"

interface BackendError {
  code: string
  message: string
  field?: string
}

export interface BackendErrorBody {
  error: BackendError
}

export function mapApiError(error: unknown): never {
  if (error instanceof AppError) {
    throw error
  }

  if (isAxiosError<BackendErrorBody>(error)) {
    const axiosError: AxiosError<BackendErrorBody> = error
    const data: BackendErrorBody | undefined = axiosError.response?.data
    const status: number = axiosError.response?.status ?? 500

    if (data?.error) {
      const { code, message, field }: BackendError = data.error

      switch (status) {
        case 400:
        case 422:
          throw new ValidationError({ message, field, cause: error })
        case 404:
          throw new NotFoundError(message, error)
        case 409:
          throw new ConflictError(message, error)
        default:
          throw new ApiError(code, message, status, field, error)
      }
    }
  }

  throw error
}
