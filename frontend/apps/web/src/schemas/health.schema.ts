import { z } from "zod"

export const HealthResponseSchema = z.object({
  status: z.string(),
})
