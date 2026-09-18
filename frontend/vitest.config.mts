// Vitest configuration for component/unit/page tests (Playwright E2E lives separately in the
// top-level e2e/ directory and is not run by this config).
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  // Native Vite `resolve.tsconfigPaths` reuses the `@/*` alias from tsconfig.json instead of
  // duplicating it here; `react()` enables JSX/Fast Refresh transforms for .tsx test files.
  plugins: [react()],
  resolve: {
    tsconfigPaths: true,
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    css: true,
    // Matches the axios client's baseURL to MSW's mocked origin so every `@/app/_services/*` call
    // resolves to a handler instead of an unset base URL. Kept as a literal (not imported from
    // src/test/constants.ts) so Vite's config loader — which runs this file directly under
    // Node rather than through its usual transform pipeline — doesn't need to load a second
    // project file; must stay in sync with TEST_API_URL there.
    env: {
      NEXT_PUBLIC_API_URL: "http://localhost/api",
    },
    exclude: ["**/node_modules/**", "**/.next/**", "../e2e/**"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary", "html"],
      reportsDirectory: "./coverage",
      exclude: [
        "**/node_modules/**",
        "**/.next/**",
        "**/*.config.*",
        "**/*.d.ts",
        "src/test/**",
        "src/app/_types/types.ts",
        "next-env.d.ts",
      ],
    },
  },
});
