// Global Vitest setup (see vitest.config.ts's `test.setupFiles`). Runs once per test file:
// extends `expect` with jest-dom + axe matchers, starts the shared MSW mock server for the
// duration of the suite, and resets per-test state (RTL's rendered DOM, MSW handler overrides,
// and the `_lib/auth.ts` access-token singleton) so tests can't leak into one another.
import "@testing-library/jest-dom/vitest";
// vitest-axe@0.1.0's own `extend-expect` entry point is broken for this project's Vitest version
// (its runtime JS is an empty no-op build, and its ambient types target an incompatible, older
// `Assertion` shape — see src/test/vitest-axe.d.ts for the replacement types). Register the actual
// matcher implementation directly from `vitest-axe/matchers` instead.
import * as axeMatchers from "vitest-axe/matchers";
import { afterAll, afterEach, beforeAll, expect } from "vitest";
import { cleanup } from "@testing-library/react";
import { setAccessToken } from "@/app/_lib/auth";
import { server } from "./msw/server";

expect.extend(axeMatchers);

// jsdom doesn't implement ResizeObserver, but react-datepicker's time-column (used by
// EventFormDateTimePicker) observes its own height on mount. A minimal no-op stub is enough since
// tests don't assert on resize-driven behavior.
if (typeof globalThis.ResizeObserver === "undefined") {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}

// `onUnhandledRequest: "error"` means a request that doesn't match any handler fails the test
// immediately (instead of silently hitting the real network or resolving with a MSW warning),
// which surfaces missing/mistyped handlers as loud test failures rather than flaky timeouts.
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));

afterEach(() => {
  // A `server.use(...)` override in one test must not bleed into the next.
  server.resetHandlers();
  // Unmount any RTL-rendered trees so effects/timers from one test don't keep running (and
  // potentially call `setState` on unmounted components) during the next test.
  cleanup();
  // `_lib/auth.ts` intentionally keeps the access token as a module-level singleton (it needs to
  // survive re-renders). Because Vitest reuses the module cache across tests in the same file,
  // that state would otherwise leak between tests — reset it so every test starts logged out.
  setAccessToken(null);
});

afterAll(() => server.close());
