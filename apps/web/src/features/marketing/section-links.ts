import { useRouterState } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";

/**
 * Resolves a landing-page section anchor for the page we are actually on: an
 * in-page jump while on `/`, and a `LANDING` + fragment link from anywhere else.
 *
 * The ids (`#levels`, `#voice`) exist only on the landing page, so a bare
 * fragment is a dead link on every other marketing page — which is exactly what
 * happened to the footer's Practice and Levels columns once About and Pricing
 * existed. Both the header and the footer resolve their anchors through here so
 * the two cannot drift apart.
 */
export function useSectionHref() {
	const pathname = useRouterState({
		select: (state) => state.location.pathname,
	});

	return (anchor: string) =>
		pathname === APP_ROUTES.LANDING ? anchor : APP_ROUTES.LANDING + anchor;
}
