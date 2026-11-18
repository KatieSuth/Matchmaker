/**
 * Matchmaker edge Worker: path-routes to Cloud Run, injects origin-verify, and
 * authenticates to Cloud Run IAM with a Google ID token.
 *
 * Env bindings (Terraform):
 *   API_ORIGIN, FRONTEND_ORIGIN, ORIGIN_VERIFY_SECRET,
 *   GCP_INVOKER_KEY (base64 service-account key JSON for the edge invoker SA)
 */

const GOOGLE_TOKEN_URL = "https://oauth2.googleapis.com/token";

// Google ID tokens live 1h; refresh early so an in-flight request never carries an expired one.
const TOKEN_REFRESH_SKEW_SECONDS = 300;

// Per-isolate caches. Entries hold a promise so concurrent cold requests share one exchange.
let invokerCredentials = null;
const idTokenCache = new Map(); // audience -> { promise, exp }

/** base64url encodes a string or byte array without padding. */
function base64url(input) {
  const bytes = typeof input === "string" ? new TextEncoder().encode(input) : new Uint8Array(input);
  let binary = "";
  for (const b of bytes) binary += String.fromCharCode(b);
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

/** Decodes the SA key binding once per isolate and imports its RSA private key for RS256 signing. */
async function getInvokerCredentials(env) {
  if (invokerCredentials) return invokerCredentials;

  const key = JSON.parse(atob(env.GCP_INVOKER_KEY));
  const pem = key.private_key
    .replace(/-----(BEGIN|END) PRIVATE KEY-----/g, "")
    .replace(/\s+/g, "");
  const der = Uint8Array.from(atob(pem), (c) => c.charCodeAt(0));
  const signingKey = await crypto.subtle.importKey(
    "pkcs8",
    der,
    { name: "RSASSA-PKCS1-v1_5", hash: "SHA-256" },
    false,
    ["sign"],
  );

  invokerCredentials = { email: key.client_email, signingKey };
  return invokerCredentials;
}

/**
 * Exchanges a self-signed service-account JWT for a Google ID token whose audience is the
 * Cloud Run service URL. Returns { token, exp } (exp in epoch seconds).
 */
async function fetchIdToken(env, audience) {
  const { email, signingKey } = await getInvokerCredentials(env);
  const now = Math.floor(Date.now() / 1000);

  const header = base64url(JSON.stringify({ alg: "RS256", typ: "JWT" }));
  const claims = base64url(
    JSON.stringify({
      iss: email,
      sub: email,
      aud: GOOGLE_TOKEN_URL,
      target_audience: audience,
      iat: now,
      exp: now + 3600,
    }),
  );
  const signature = await crypto.subtle.sign(
    "RSASSA-PKCS1-v1_5",
    signingKey,
    new TextEncoder().encode(`${header}.${claims}`),
  );
  const assertion = `${header}.${claims}.${base64url(signature)}`;

  const res = await fetch(GOOGLE_TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "urn:ietf:params:oauth:grant-type:jwt-bearer",
      assertion,
    }),
  });
  if (!res.ok) {
    throw new Error(`id token exchange failed: ${res.status}`);
  }

  const { id_token: token } = await res.json();
  // Trust the token's own exp over our requested lifetime.
  const payload = JSON.parse(atob(token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")));
  return { token, exp: payload.exp };
}

/** Returns a cached ID token for audience, refreshing it shortly before expiry. */
async function getIdToken(env, audience) {
  const now = Math.floor(Date.now() / 1000);
  const cached = idTokenCache.get(audience);
  if (cached && (cached.exp === null || now < cached.exp - TOKEN_REFRESH_SKEW_SECONDS)) {
    return (await cached.promise).token;
  }

  // exp is unknown until the exchange resolves; null marks the entry as in-flight.
  const entry = { promise: fetchIdToken(env, audience), exp: null };
  idTokenCache.set(audience, entry);
  try {
    const result = await entry.promise;
    entry.exp = result.exp;
    return result.token;
  } catch (err) {
    if (idTokenCache.get(audience) === entry) idTokenCache.delete(audience);
    throw err;
  }
}

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

    // Cloud Run checks X-Serverless-Authorization for IAM and leaves Authorization
    // (the app's own JWT) untouched. set() also overwrites any client-supplied value.
    try {
      headers.set("X-Serverless-Authorization", `Bearer ${await getIdToken(env, origin)}`);
    } catch (err) {
      console.error("cloud run id token unavailable", err);
      return new Response("Bad Gateway", { status: 502 });
    }

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
