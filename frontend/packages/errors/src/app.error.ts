export interface AppErrorOptions {
  code: string;
  message: string;
  status: number;
  field?: string | undefined;
  cause?: unknown;
}

export class AppError extends Error {
  readonly code: string;
  readonly status: number;
  readonly field: string | undefined;

  public constructor({ code, message, status, field, cause }: AppErrorOptions) {
    super(message, { cause });

    this.name = new.target.name;
    this.code = code;
    this.status = status;
    this.field = field;
  }
}
