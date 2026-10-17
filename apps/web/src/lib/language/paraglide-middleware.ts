import { createMiddleware } from "@tanstack/react-start";
import { paraglideMiddleware } from "#/paraglide/server.js";

/**
 * SSR locale via Paraglide's own server middleware: detects the request
 * locale (cookie strategy → PARAGLIDE_LOCALE, else baseLocale vi) and scopes
 * it with AsyncLocalStorage. No `url` strategy is configured, so the request
 * passes through untouched — only the locale context is set.
 */
export const paraglideRequestMiddleware = createMiddleware({
	type: "request",
}).server(({ request, next }) => paraglideMiddleware(request, () => next()));
