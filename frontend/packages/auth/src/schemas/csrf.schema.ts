import { z } from "zod/v4";

export const CsrfResponseSchema = z.object({
  csrfToken: z.string().min(1),
});

export type CsrfResponse = z.infer<typeof CsrfResponseSchema>;
