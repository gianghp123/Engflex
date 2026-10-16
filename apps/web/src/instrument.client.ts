import * as Sentry from "@sentry/tanstackstart-react";
import { env } from "#/env";

// The browser DSN is public by nature (it ships in the client bundle), but
// it is still configured per environment — never hardcoded — so the key
// can rotate and dev builds stay out of the production issue stream.
// Empty means disabled, matching the api/voice convention.
const dsn = env.VITE_SENTRY_DSN;
if (dsn) {
	Sentry.init({
		dsn,
		integrations: [Sentry.replayIntegration()],
		tracesSampleRate: 1.0,
		replaysSessionSampleRate: 0.1,
		replaysOnErrorSampleRate: 1.0,
	});
}
