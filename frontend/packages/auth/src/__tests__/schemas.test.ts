import { describe, it, expect } from "vitest";
import { AuthUserSchema } from "../schemas/user.schema";
import { AuthSessionSchema } from "../schemas/session.schema";
import { AuthContextSchema } from "../schemas/context.schema";
import { AuthSessionResponseSchema } from "../schemas/auth-response.schema";
import { CsrfResponseSchema } from "../schemas/csrf.schema";

const validUuid = "550e8400-e29b-41d4-a716-446655440000";
const validDatetime = "2026-07-19T12:00:00Z";

describe("AuthUserSchema", () => {
  it("accepts valid user", () => {
    const result = AuthUserSchema.parse({
      id: validUuid,
      email: "user@example.com",
      displayName: "User Name",
      emailVerified: true,
    });
    expect(result.email).toBe("user@example.com");
  });

  it("rejects invalid email", () => {
    expect(() =>
      AuthUserSchema.parse({
        id: validUuid,
        email: "not-an-email",
        displayName: "User",
        emailVerified: true,
      }),
    ).toThrow();
  });

  it("rejects empty displayName", () => {
    expect(() =>
      AuthUserSchema.parse({
        id: validUuid,
        email: "user@example.com",
        displayName: "",
        emailVerified: true,
      }),
    ).toThrow();
  });
});

describe("AuthSessionSchema", () => {
  it("accepts valid session", () => {
    const result = AuthSessionSchema.parse({
      id: validUuid,
      clientId: "pdv-web",
      createdAt: validDatetime,
      idleExpiresAt: validDatetime,
      absoluteExpiresAt: validDatetime,
    });
    expect(result.clientId).toBe("pdv-web");
  });

  it("rejects invalid datetime", () => {
    expect(() =>
      AuthSessionSchema.parse({
        id: validUuid,
        clientId: "pdv-web",
        createdAt: "not-a-date",
        idleExpiresAt: validDatetime,
        absoluteExpiresAt: validDatetime,
      }),
    ).toThrow();
  });
});

describe("AuthContextSchema", () => {
  it("accepts identity context", () => {
    const result = AuthContextSchema.parse({
      kind: "identity",
      membershipId: null,
      organization: null,
      store: null,
      roles: [],
      scopes: [],
    });
    expect(result.kind).toBe("identity");
  });

  it("accepts organization context", () => {
    const result = AuthContextSchema.parse({
      kind: "organization",
      membershipId: validUuid,
      organization: { id: validUuid, name: "Org", slug: "org" },
      store: null,
      roles: ["admin"],
      scopes: ["organization.read"],
    });
    expect(result.kind).toBe("organization");
  });

  it("accepts store context", () => {
    const result = AuthContextSchema.parse({
      kind: "store",
      membershipId: validUuid,
      organization: { id: validUuid, name: "Org", slug: "org" },
      store: { id: validUuid, code: "STORE01", name: "Store 1" },
      roles: ["cashier"],
      scopes: ["sales.create"],
    });
    expect(result.kind).toBe("store");
  });

  it("rejects identity with organization", () => {
    expect(() =>
      AuthContextSchema.parse({
        kind: "identity",
        membershipId: validUuid,
        organization: { id: validUuid, name: "Org", slug: "org" },
        store: null,
        roles: [],
        scopes: [],
      }),
    ).toThrow();
  });

  it("rejects organization without membershipId", () => {
    expect(() =>
      AuthContextSchema.parse({
        kind: "organization",
        membershipId: null,
        organization: { id: validUuid, name: "Org", slug: "org" },
        store: null,
        roles: [],
        scopes: [],
      }),
    ).toThrow();
  });

  it("rejects store without store field", () => {
    expect(() =>
      AuthContextSchema.parse({
        kind: "store",
        membershipId: validUuid,
        organization: { id: validUuid, name: "Org", slug: "org" },
        store: null,
        roles: [],
        scopes: [],
      }),
    ).toThrow();
  });

  it("rejects store without organization", () => {
    expect(() =>
      AuthContextSchema.parse({
        kind: "store",
        membershipId: validUuid,
        organization: null,
        store: { id: validUuid, code: "STORE01", name: "Store 1" },
        roles: [],
        scopes: [],
      }),
    ).toThrow();
  });
});

describe("AuthSessionResponseSchema", () => {
  it("accepts valid full response", () => {
    const result = AuthSessionResponseSchema.parse({
      accessToken: "jwt-token",
      tokenType: "Bearer",
      expiresIn: 300,
      user: {
        id: validUuid,
        email: "user@example.com",
        displayName: "User",
        emailVerified: true,
      },
      session: {
        id: validUuid,
        clientId: "pdv-web",
        createdAt: validDatetime,
        idleExpiresAt: validDatetime,
        absoluteExpiresAt: validDatetime,
      },
      context: {
        kind: "identity",
        membershipId: null,
        organization: null,
        store: null,
        roles: [],
        scopes: [],
      },
    });
    expect(result.tokenType).toBe("Bearer");
    expect(result.expiresIn).toBe(300);
  });

  it("rejects empty accessToken", () => {
    expect(() =>
      AuthSessionResponseSchema.parse({
        accessToken: "",
        tokenType: "Bearer",
        expiresIn: 300,
        user: {
          id: validUuid,
          email: "a@b.com",
          displayName: "U",
          emailVerified: true,
        },
        session: {
          id: validUuid,
          clientId: "pdv-web",
          createdAt: validDatetime,
          idleExpiresAt: validDatetime,
          absoluteExpiresAt: validDatetime,
        },
        context: {
          kind: "identity",
          membershipId: null,
          organization: null,
          store: null,
          roles: [],
          scopes: [],
        },
      }),
    ).toThrow();
  });

  it("rejects non-Bearer tokenType", () => {
    expect(() =>
      AuthSessionResponseSchema.parse({
        accessToken: "token",
        tokenType: "MAC",
        expiresIn: 300,
        user: {
          id: validUuid,
          email: "a@b.com",
          displayName: "U",
          emailVerified: true,
        },
        session: {
          id: validUuid,
          clientId: "pdv-web",
          createdAt: validDatetime,
          idleExpiresAt: validDatetime,
          absoluteExpiresAt: validDatetime,
        },
        context: {
          kind: "identity",
          membershipId: null,
          organization: null,
          store: null,
          roles: [],
          scopes: [],
        },
      }),
    ).toThrow();
  });

  it("rejects zero expiresIn", () => {
    expect(() =>
      AuthSessionResponseSchema.parse({
        accessToken: "token",
        tokenType: "Bearer",
        expiresIn: 0,
        user: {
          id: validUuid,
          email: "a@b.com",
          displayName: "U",
          emailVerified: true,
        },
        session: {
          id: validUuid,
          clientId: "pdv-web",
          createdAt: validDatetime,
          idleExpiresAt: validDatetime,
          absoluteExpiresAt: validDatetime,
        },
        context: {
          kind: "identity",
          membershipId: null,
          organization: null,
          store: null,
          roles: [],
          scopes: [],
        },
      }),
    ).toThrow();
  });
});

describe("CsrfResponseSchema", () => {
  it("accepts valid response", () => {
    const result = CsrfResponseSchema.parse({ csrfToken: "base64-token" });
    expect(result.csrfToken).toBe("base64-token");
  });

  it("rejects empty token", () => {
    expect(() => CsrfResponseSchema.parse({ csrfToken: "" })).toThrow();
  });
});
