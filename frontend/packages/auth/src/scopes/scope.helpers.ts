import type { AuthContext } from "../schemas/context.schema";
import type { Scope } from "../types/scope.types";

export function hasScope(context: AuthContext, scope: Scope): boolean {
  return context.scopes.includes(scope);
}

export function hasAllScopes(
  context: AuthContext,
  scopes: readonly Scope[],
): boolean {
  if (scopes.length === 0) return true;
  const scopeSet = new Set(context.scopes);
  return scopes.every((s) => scopeSet.has(s));
}

export function hasAnyScope(
  context: AuthContext,
  scopes: readonly Scope[],
): boolean {
  if (scopes.length === 0) return false;
  const scopeSet = new Set(context.scopes);
  return scopes.some((s) => scopeSet.has(s));
}
