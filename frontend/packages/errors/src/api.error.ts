import { AppError } from "./app.error";

export class ApiError extends AppError {
  public constructor(
    code: string,
    message: string,
    status: number,
    field?: string | undefined,
    cause?: unknown,
  ) {
    super({ code, message, status, field, cause });
  }
}
