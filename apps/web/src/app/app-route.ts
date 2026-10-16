/** Centralized app route URLs. Use these instead of hardcoding paths. */
export const APP_ROUTES = {
	HOME: "/",
	ABOUT: "/about",
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
	ADMIN: {
		OVERVIEW: "/admin",
		VOCABULARY: "/admin/vocabulary",
		VIDEOS: "/admin/videos",
		SCENARIOS: "/admin/scenarios",
		LESSONS: "/admin/lessons",
		USERS: "/admin/users",
	},
} as const;
