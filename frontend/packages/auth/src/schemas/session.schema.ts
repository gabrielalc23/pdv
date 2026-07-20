import { z } from "zod/v4";

export const AuthSessionSchema = z.object({
  id: z.string().uuid(),
  clientId: z.string().min(1).max(50),
  createdAt: z.string().datetime(),
  idleExpiresAt: z.string().datetime(),
  absoluteExpiresAt: z.string().datetime(),
});

export type AuthSession = z.infer<typeof AuthSessionSchema>;
