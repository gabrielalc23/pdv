export type Left<L> = {
  readonly type: "left"
  readonly value: L
}

export type Right<R> = {
  readonly type: "right"
  readonly value: R
}

export type Either<L, R> = Left<L> | Right<R>
