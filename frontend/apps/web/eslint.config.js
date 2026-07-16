import reactRefresh from "eslint-plugin-react-refresh"
import { defineConfig } from "eslint/config"
import { createWorkspaceEslintConfig } from "@pdv/eslint"

export default defineConfig([
  ...createWorkspaceEslintConfig({ tsconfigRootDir: import.meta.dirname }),

  reactRefresh.configs.vite,

  {
    files: ["src/routes/**/*.{ts,tsx}", "src/**/__tests__/**/*.{ts,tsx}"],
    rules: {
      "react-refresh/only-export-components": "off",
    },
  },
])
