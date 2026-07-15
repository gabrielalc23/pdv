import { AppError } from "./app.error"

export class UnauthorizedError extends AppError {
  public constructor(
    message: string = "Authentication is required",
    cause?: unknown,
  ) {
    super({
      code: "unauthorized",
      message,
      status: 401,
      cause,
    })
  }
}