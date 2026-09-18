// Shared constants for the Vitest / MSW suite. Mirrors `@/app/_lib/constants.ts` for app-level
// values: keep test-only literals here instead of scattering them through handlers and specs.
//
// TEST_API_URL must stay in sync with the `NEXT_PUBLIC_API_URL` literal in vitest.config.mts
// (which injects this same value as the axios client's baseURL). It's duplicated as a literal
// there rather than imported to avoid Vite's config loader warning about loading project ESM
// files through its Node-native (CommonJS-style) config-loading path.

/** Fake API origin used by every MSW handler and by axios in the test environment. */
export const TEST_API_URL = "http://localhost/api";
