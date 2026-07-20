import { z } from "zod/v4";

const ContextKind = z.enum(["identity", "organization", "store"]);

const OrganizationRefSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  slug: z.string().min(1),
});

const StoreRefSchema = z.object({
  id: z.string().uuid(),
  code: z.string().min(1),
  name: z.string().min(1),
});

export const AuthContextSchema = z
  .object({
    kind: ContextKind,
    membershipId: z.string().uuid().nullable(),
    organization: OrganizationRefSchema.nullable(),
    store: StoreRefSchema.nullable(),
    roles: z.array(z.string()).default([]),
    scopes: z.array(z.string()).default([]),
  })
  .refine(
    (ctx) => {
      if (ctx.kind === "identity") {
        return (
          ctx.membershipId === null &&
          ctx.organization === null &&
          ctx.store === null
        );
      }
      return true;
    },
    {
      message:
        "Identity context must have null membershipId, organization, and store",
    },
  )
  .refine(
    (ctx) => {
      if (ctx.kind === "organization") {
        return (
          ctx.membershipId !== null &&
          ctx.organization !== null &&
          ctx.store === null
        );
      }
      return true;
    },
    {
      message:
        "Organization context must have membershipId and organization, but no store",
    },
  )
  .refine(
    (ctx) => {
      if (ctx.kind === "store") {
        return (
          ctx.membershipId !== null &&
          ctx.organization !== null &&
          ctx.store !== null
        );
      }
      return true;
    },
    {
      message: "Store context must have membershipId, organization, and store",
    },
  );

export type AuthContext = z.infer<typeof AuthContextSchema>;
