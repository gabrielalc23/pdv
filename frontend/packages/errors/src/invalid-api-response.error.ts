import { AppError } from "./app.error";

export class InvalidApiResponseError extends AppError {
  public constructor(path: string, cause?: unknown) {
    super({
      code: "invalid_api_response",
      message: `The API response for ${path} does not match the expected schema`,
      status: 502,
      cause,
    });
  }
}
