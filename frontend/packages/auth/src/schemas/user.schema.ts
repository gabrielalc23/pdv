import { z } from "zod/v4";

export const AuthUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email().max(320),
  displayName: z.string().min(1).max(150),
  emailVerified: z.boolean(),
});

export type AuthUser = z.infer<typeof AuthUserSchema>;
