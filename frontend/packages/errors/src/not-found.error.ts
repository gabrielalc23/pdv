import { AppError } from "./app.error"

export class NotFoundError extends AppError {
  public constructor(message: string = "Resource not found", cause?: unknown) {
    super({
      code: "not_found",
      message,
      status: 404,
      cause,
    })
  }
}
