export interface AppErrorOptions {
  code: string
  message: string
  status: number
  field?: string
  cause?: unknown
}

export class AppError extends Error {
  readonly code: string
  readonly status: number
  readonly field?: string

  public constructor({
    code,
    message,
    status,
    field,
    cause,
  }: AppErrorOptions) {
    super(message, { cause })

    this.name = new.target.name
    this.code = code
    this.status = status
    this.field = field
  }
}