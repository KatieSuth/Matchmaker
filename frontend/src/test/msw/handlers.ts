// Combined default MSW handlers, registered by both the Vitest node server (src/test/setup.ts)
// and the optional browser worker (src/test/msw/browser.ts). Individual tests override a
// specific route for that test only via `server.use(...)` — see MSW's docs on runtime overrides.
import { authHandlers } from "./handlers/auth";
import { eventsHandlers } from "./handlers/events";
import { gamesHandlers } from "./handlers/games";
import { usersHandlers } from "./handlers/users";

export const handlers = [...authHandlers, ...eventsHandlers, ...gamesHandlers, ...usersHandlers];
