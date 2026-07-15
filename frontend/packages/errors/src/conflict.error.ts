import { AppError } from "./app.error"

export class ConflictError extends AppError {
  public constructor(
    message: string = "The resource conflicts with the current state",
    cause?: unknown,
  ) {
    super({
      code: "conflict",
      message,
      status: 409,
      cause,
    })
  }
}