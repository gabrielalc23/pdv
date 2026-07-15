import { AppError } from "./app.error"

export class ForbiddenError extends AppError {
  public constructor(
    message: string = "You do not have permission to perform this action",
    cause?: unknown,
  ) {
    super({
      code: "forbidden",
      message,
      status: 403,
      cause,
    })
  }
}