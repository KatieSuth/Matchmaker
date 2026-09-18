// Node-side MSW request-interception server, used by Vitest (see src/test/setup.ts). Wired up
// with `onUnhandledRequest: "error"` there so an un-mocked call fails the test loudly instead of
// hitting the real network or silently resolving with a confusing error.
import { setupServer } from "msw/node";
import { handlers } from "./handlers";

export const server = setupServer(...handlers);
