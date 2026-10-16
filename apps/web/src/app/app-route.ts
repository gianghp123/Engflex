/** Centralized app route URLs. Use these instead of hardcoding paths. */
export const APP_ROUTES = {
	/** Public landing page — the front door for signed-out visitors. */
	LANDING: "/",
	HOME: "/home",
	ABOUT: "/about",
	PRICING: "/pricing",
	/**
	 * Section anchors on the landing page. The shared marketing nav turns these
	 * into an in-page jump while it is on `/`, and into a cross-page link
	 * (`LANDING` + fragment) on every other marketing page — the ids only exist
	 * on the landing page, so a bare fragment elsewhere is a dead link.
	 */
	LANDING_SECTIONS: {
		LEVELS: "#levels",
		VOICE: "#voice",
	},
	ONBOARDING: "/onboarding",
	LESSONS: {
		LIST: "/lessons",
		DETAIL: "/lessons/$slug",
		PRACTICE: "/lessons/$slug/practice",
	},
	VOCABULARY: {
		LIST: "/vocabulary",
		DETAIL: "/vocabulary/$itemId",
	},
	VOICE: {
		LIST: "/voice",
		SCENARIOS: "/voice/scenarios",
		SCENARIO: "/voice/scenarios/$scenarioId",
		ROOM: "/voice/room/$conversationId",
		SIGNAL_START: "/api/voice/$conversationId/start",
		SIGNAL_OFFER: "/api/voice/$conversationId/offer",
	},
	VIDEOS: {
		LIST: "/videos",
		DETAIL: "/videos/$videoId",
		DICTATION: "/videos/$videoId/dictation",
		SHADOWING: "/videos/$videoId/shadowing",
		CATEGORY: "/videos/categories/$categorySlug",
	},
	/**
	 * Clerk's catch-all pages. The file routes are `/sign-in/$` splats, so a
	 * typed `to` needs the splat id — but `redirect({ href })` bypasses route
	 * resolution entirely and must carry a plain URL, or the browser navigates
	 * to a literal `$`. These values are therefore href-only.
	 */
	AUTH: {
		SIGN_IN: "/sign-in",
		SIGN_UP: "/sign-up",
	},
	/** Dedicated status pages, both rendered by `components/common/error-page`. */
	ERROR: {
		NOT_FOUND: "/not-found",
		SERVER: "/error",
	},
} as const;
