import { tanstackConfig } from "@tanstack/eslint-config"

interface WorkspaceEslintConfigOptions {
  tsconfigRootDir: string
  tsconfigProject?: string | string[]
}

export function createWorkspaceEslintConfig({
  tsconfigRootDir,
  tsconfigProject = ["./tsconfig.app.json", "./tsconfig.node.json"],
}: WorkspaceEslintConfigOptions) {
  return [
    ...tanstackConfig,
    {
      files: ["**/*.{ts,tsx}"],
      languageOptions: {
        parserOptions: {
          project: tsconfigProject,
          tsconfigRootDir,
        },
      },
    },
    {
      rules: {
        semi: ["error", "never"],
        "import/no-cycle": "off",
        "import/order": "off",
        "sort-imports": "off",
        "@typescript-eslint/array-type": "off",
        "@typescript-eslint/require-await": "off",
        "pnpm/json-enforce-catalog": "off",
      },
    },
    {
      ignores: ["eslint.config.js", "prettier.config.js"],
    },
  ]
}
