// Route guard: uses the auth_session cookie to redirect; API routes still require a valid JWT.
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function proxy(request: NextRequest) {
    /* note: "isAuthenticated" is just a lightweight flag set on login & token refresh for the
     * sake of middleware redirects. If someone is trying to get clever and change this cookie,
     * the middleware will let them go to whatever page, but the API will still fail their request.
     * If they didn't want the app to "break" they shouldn't have manually changed this value anyway.
     */
    const isAuthenticated = request.cookies.has("auth_session");
    const { pathname } = request.nextUrl;

    // If no refresh token, always redirect to "/"
    if (!isAuthenticated && pathname !== "/") {
        return NextResponse.redirect(new URL("/", request.url));
    }

    // If refresh token exists and they hit "/", send to "/events"
    if (isAuthenticated && pathname === "/") {
        return NextResponse.redirect(new URL("/my_events", request.url));
    }

    return NextResponse.next();
}

export const config = {
    // Run middleware on all routes except Next.js internals and static files
    matcher: ["/((?!_next/static|_next/image|favicon.ico|auth/callback|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.svg|.*\\.ico|.*\\.webp|.*\\.gif).*)"],
    //matcher: ["/:path*"],
};