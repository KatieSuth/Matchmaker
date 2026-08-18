// Route guard: uses the auth_session cookie to redirect; API routes still require a valid JWT.
// When ORIGIN_VERIFY_SECRET is set, requires matching X-Origin-Verify (injected by Cloudflare Worker).
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { isAllowedPostLoginPath } from "@/app/_lib/postLoginRedirect";

const ORIGIN_VERIFY_HEADER = "x-origin-verify";

function originVerifyForbidden(): NextResponse {
    return new NextResponse("Forbidden", { status: 403 });
}

export function proxy(request: NextRequest) {
    const secret = process.env.ORIGIN_VERIFY_SECRET;
    const { pathname } = request.nextUrl;
    const isHealthPath = pathname === "/health";
    const isAuthCallbackPath = pathname === "/auth/callback";
    const isAboutPath = pathname === "/about";

    // Cloud Run / Docker probes hit /health without the Worker header.
    if (secret && !isHealthPath) {
        const got = request.headers.get(ORIGIN_VERIFY_HEADER);
        if (got !== secret) {
            return originVerifyForbidden();
        }
    }

    /* note: "isAuthenticated" is just a lightweight flag set on login & token refresh for the
     * sake of middleware redirects. If someone is trying to get clever and change this cookie,
     * the middleware will let them go to whatever page, but the API will still fail their request.
     * If they didn't want the app to "break" they shouldn't have manually changed this value anyway.
     */
    const isAuthenticated = request.cookies.has("auth_session");

    // Discord OAuth lands on /auth/callback before auth_session exists — do not bounce to /.
    // /about is a public marketing page for both auth states.
    // Logged-out users otherwise only see the landing page; preserve deep links (?next=/event/...)
    if (!isAuthenticated && pathname !== "/" && !isHealthPath && !isAuthCallbackPath && !isAboutPath) {
        const loginUrl = new URL("/", request.url);
        loginUrl.searchParams.set("next", pathname + request.nextUrl.search);
        return NextResponse.redirect(loginUrl);
    }

    // Logged-in users on "/" go to My Events, or to ?next= when it is a safe in-app path
    if (isAuthenticated && pathname === "/") {
        const nextParam = request.nextUrl.searchParams.get("next");
        if (nextParam && isAllowedPostLoginPath(nextParam)) {
            return NextResponse.redirect(new URL(nextParam, request.url));
        }
        return NextResponse.redirect(new URL("/my_events", request.url));
    }

    return NextResponse.next();
}

export const config = {
    // Run on all routes except Next.js internals and static assets.
     // Negative lookahead = paths this proxy does NOT run on.
    // /auth/callback and /about are intentionally NOT listed here so origin-verify still runs;
    // the unauthenticated redirect above skips those paths so OAuth can finish and /about stays public.
    // /health is listed here so probes hit public/health without this proxy.
    matcher: [
        "/((?!_next/static|_next/image|favicon.ico|health|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.svg|.*\\.ico|.*\\.webp|.*\\.gif).*)",
    ],
};
