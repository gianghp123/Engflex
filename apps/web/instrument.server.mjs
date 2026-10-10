import * as Sentry from "@sentry/tanstackstart-react";

// Server-side Sentry init: env-driven and disabled when the DSN is empty —
// the same contract as instrument.client.ts. Never hardcode a DSN here.
const dsn = process.env.VITE_SENTRY_DSN;
if (dsn) {
  Sentry.init({
    dsn,
    tracesSampleRate: 1.0,
  });
}
