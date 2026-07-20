import { z } from "zod/v4";
import { AuthUserSchema } from "./user.schema";
import { AuthSessionSchema } from "./session.schema";
import { AuthContextSchema } from "./context.schema";

export const AuthSessionResponseSchema = z.object({
  accessToken: z.string().min(1),
  tokenType: z.literal("Bearer"),
  expiresIn: z.number().int().positive(),
  user: AuthUserSchema,
  session: AuthSessionSchema,
  context: AuthContextSchema,
});

export type AuthSessionResponse = z.infer<typeof AuthSessionResponseSchema>;
