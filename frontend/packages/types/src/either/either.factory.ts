import type {
  Left,
  Right,
} from './either.type'

export const left: <L>(value: L) => Left<L> = <L>(value: L): Left<L> => ({
  type: 'left',
  value,
})

export const right: <R>(value: R) => Right<R> = <R>(value: R): Right<R> => ({
  type: 'right',
  value,
})