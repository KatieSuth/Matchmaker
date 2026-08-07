// Sentry (optional): initializes only when SENTRY_DSN is set (Cloud Run / prod).
export async function register() {
  const dsn = process.env.SENTRY_DSN;
  if (!dsn) {
    return;
  }
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const Sentry = await import("@sentry/nextjs");
    Sentry.init({
      dsn,
      tracesSampleRate: 0,
    });
  }
}
