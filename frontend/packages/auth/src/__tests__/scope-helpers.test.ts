import { describe, it, expect } from "vitest";
import { hasScope, hasAllScopes, hasAnyScope } from "../scopes/scope.helpers";
import type { AuthContext } from "../schemas/context.schema";

const identityContext: AuthContext = {
  kind: "identity",
  membershipId: null,
  organization: null,
  store: null,
  roles: [],
  scopes: [],
};

const orgContext: AuthContext = {
  kind: "organization",
  membershipId: "550e8400-e29b-41d4-a716-446655440000",
  organization: {
    id: "550e8400-e29b-41d4-a716-446655440001",
    name: "Org",
    slug: "org",
  },
  store: null,
  roles: ["admin"],
  scopes: ["organization.read", "products.read", "members.read"],
};

const storeContext: AuthContext = {
  kind: "store",
  membershipId: "550e8400-e29b-41d4-a716-446655440000",
  organization: {
    id: "550e8400-e29b-41d4-a716-446655440001",
    name: "Org",
    slug: "org",
  },
  store: {
    id: "550e8400-e29b-41d4-a716-446655440002",
    code: "STORE01",
    name: "Store 1",
  },
  roles: ["cashier"],
  scopes: ["catalog.read", "sales.create", "sales.checkout"],
};

describe("hasScope", () => {
  it("returns true for present scope", () => {
    expect(hasScope(orgContext, "organization.read")).toBe(true);
  });

  it("returns false for absent scope", () => {
    expect(hasScope(orgContext, "stores.create")).toBe(false);
  });

  it("returns false for identity without scopes", () => {
    expect(hasScope(identityContext, "organization.read")).toBe(false);
  });
});

describe("hasAllScopes", () => {
  it("returns true when all present", () => {
    expect(
      hasAllScopes(orgContext, ["organization.read", "products.read"]),
    ).toBe(true);
  });

  it("returns false when one missing", () => {
    expect(
      hasAllScopes(orgContext, ["organization.read", "stores.create"]),
    ).toBe(false);
  });

  it("returns true for empty array", () => {
    expect(hasAllScopes(orgContext, [])).toBe(true);
  });

  it("works with store scopes", () => {
    expect(hasAllScopes(storeContext, ["catalog.read", "sales.create"])).toBe(
      true,
    );
  });
});

describe("hasAnyScope", () => {
  it("returns true when one matches", () => {
    expect(
      hasAnyScope(orgContext, [
        "organization.read" as any,
        "nonexistent" as any,
      ]),
    ).toBe(true);
  });

  it("returns true when all match", () => {
    expect(
      hasAnyScope(orgContext, ["organization.read", "products.read"]),
    ).toBe(true);
  });

  it("returns false for no match", () => {
    expect(hasAnyScope(orgContext, ["stores.create", "members.invite"])).toBe(
      false,
    );
  });

  it("returns false for empty array", () => {
    expect(hasAnyScope(orgContext, [])).toBe(false);
  });

  it("returns false for identity without scopes", () => {
    expect(hasAnyScope(identityContext, ["organization.read"])).toBe(false);
  });
});
