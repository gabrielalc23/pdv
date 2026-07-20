import { z } from "zod/v4";

const SourceIdSchema = z.string().uuid().or(z.string().min(8));

export const TokenUpdatedMessageSchema = z.object({
  type: z.literal("token-updated"),
  sourceId: SourceIdSchema,
  accessToken: z.string().min(1),
  expiresIn: z.number().int().positive(),
  session: z.unknown(),
});

export const LogoutMessageSchema = z.object({
  type: z.literal("logout"),
  sourceId: SourceIdSchema,
});

export const AuthLostMessageSchema = z.object({
  type: z.literal("auth-lost"),
  sourceId: SourceIdSchema,
  reason: z.string().min(1),
});

export const ContextChangedMessageSchema = z.object({
  type: z.literal("context-changed"),
  sourceId: SourceIdSchema,
  accessToken: z.string().min(1),
  expiresIn: z.number().int().positive(),
  session: z.unknown(),
});

export const AuthBroadcastMessageSchema = z.discriminatedUnion("type", [
  TokenUpdatedMessageSchema,
  LogoutMessageSchema,
  AuthLostMessageSchema,
  ContextChangedMessageSchema,
]);

export type AuthBroadcastMessageParsed = z.infer<
  typeof AuthBroadcastMessageSchema
>;
