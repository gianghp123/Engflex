/** Centralized app route URLs. Use these instead of hardcoding paths. */
export const APP_ROUTES = {
	/** Public landing page — the front door for signed-out visitors. */
	LANDING: "/",
	HOME: "/home",
	ABOUT: "/about",
	PRICING: "/pricing",
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
		/** Past sessions: the list and one session's review. */
		HISTORY: {
			LIST: "/voice/history",
			DETAIL: "/voice/history/$conversationId",
		},
		SIGNAL_START: "/api/voice/$conversationId/start",
		SIGNAL_OFFER: "/api/voice/$conversationId/offer",
	},
	VIDEOS: {
		LIST: "/videos",
		DICTATION: "/videos/$videoId/dictation",
		SHADOWING: "/videos/$videoId/shadowing",
		CATEGORY: "/videos/categories/$categorySlug",
	},
	STORIES: {
		LIST: "/stories",
		DETAIL: "/stories/$slug",
	},
	MEDIA: {
		DETAIL: "/api/media/$",
	},
	EXAMS: {
		TOEIC: {
			LIST: "/exams/toeic",
			CATEGORY: "/exams/toeic/$categoryId",
			TEST: "/exams/toeic/$categoryId/$testId",
			SESSION: "/exams/toeic/session/$id",
			RESULT: "/exams/toeic/session/$id/result",
			HISTORY: "/exams/toeic/history",
		},
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
		PREMIUM: "/admin/premium",
	},
	PAYMENTS: {
		PRICING: "/payments/pricing",
		RETURN: "/payments/vnpay/return",
	},
} as const;
