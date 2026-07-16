import { AppError } from "./app.error"

export interface ValidationErrorOptions {
  message?: string
  field?: string | undefined
  details?: Record<string, unknown> | undefined
  cause?: unknown
}

export class ValidationError extends AppError {
  public readonly details: Record<string, unknown> | undefined

  public constructor({
    message = "The provided data is invalid",
    field,
    details,
    cause,
  }: ValidationErrorOptions = {}) {
    super({
      code: "validation_error",
      message,
      status: 422,
      field,
      cause,
    })

    this.details = details
  }
}
