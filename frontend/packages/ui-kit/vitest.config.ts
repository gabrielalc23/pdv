import { defineConfig } from "vitest/config"
import react from "@vitejs/plugin-react"
import * as path from "node:path"

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "#components": path.resolve(__dirname, "src/components"),
      "#hooks": path.resolve(__dirname, "src/hooks"),
      "#lib": path.resolve(__dirname, "src/lib"),
    },
    conditions: ["import", "browser", "module", "default"],
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/__tests__/setup.ts"],
    include: ["src/**/*.test.tsx", "src/**/*.test.ts"],
    css: false,
    server: {
      deps: {
        inline: ["@base-ui/react"],
      },
    },
  },
})
