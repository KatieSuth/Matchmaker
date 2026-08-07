/**
 * Matchmaker edge Worker: path-routes to Cloud Run and injects origin-verify.
 *
 * Env bindings (Terraform):
 *   API_ORIGIN, FRONTEND_ORIGIN, ORIGIN_VERIFY_SECRET
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const isAPI = url.pathname === "/api" || url.pathname.startsWith("/api/");

    const origin = (isAPI ? env.API_ORIGIN : env.FRONTEND_ORIGIN).replace(/\/$/, "");
    const path = isAPI ? url.pathname.slice("/api".length) || "/" : url.pathname;
    const target = new URL(path + url.search, origin + "/");

    const headers = new Headers(request.headers);
    headers.set("X-Origin-Verify", env.ORIGIN_VERIFY_SECRET);
    headers.set("Host", target.host);

    const cfConnecting = request.headers.get("CF-Connecting-IP");
    if (cfConnecting) {
      headers.set("X-Forwarded-For", cfConnecting);
    }

    const init = {
      method: request.method,
      headers,
      redirect: "manual",
    };
    if (request.method !== "GET" && request.method !== "HEAD") {
      init.body = request.body;
      init.duplex = "half";
    }

    const response = await fetch(target.toString(), init);
    const out = new Response(response.body, response);
    out.headers.set("X-Content-Type-Options", "nosniff");
    out.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
    out.headers.set("X-Frame-Options", "DENY");
    out.headers.delete("X-Origin-Verify");
    return out;
  },
};
