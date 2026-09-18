// Shared E2E endpoints. Playwright talks to Caddy for the app and to the
// published API port for POST /auth/test_login (avoids TLS on the Node fetch).

export const e2eApiUrl = process.env.E2E_API_URL ?? "http://127.0.0.1:18080";
export const e2eBypassToken = process.env.E2E_BYPASS_TOKEN ?? "e2e-bypass-token";
export const testAuthBypassHeader = "X-Test-Auth-Bypass-Token";
