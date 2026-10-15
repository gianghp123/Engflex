type ApiId = string | number;

/** Backend paths only. No hardcoded URLs in services — import from here. */
export const API_ROUTES = {
	PROFILES: {
		ME: "/profiles/me",
		BY_ID: (id: ApiId) => `/profiles/${id}`,
	},
	CONTENTS: {
		LIST: "/contents",
		BY_ID: (id: ApiId) => `/contents/${id}`,
	},
	EXERCISES: {
		LIST: "/exercises",
		BY_ID: (id: ApiId) => `/exercises/${id}`,
	},
	ATTEMPTS: {
		LIST: "/attempts",
		BY_ID: (id: ApiId) => `/attempts/${id}`,
	},
	VOCABULARY: {
		ITEMS: {
			LIST: "/vocabulary-items",
			BY_ID: (id: ApiId) => `/vocabulary-items/${id}`,
		},
		LIBRARY: {
			LIST: "/user-vocabulary",
			STATS: "/user-vocabulary/stats",
			ADD: "/user-vocabulary/words",
			BY_ID: (itemId: ApiId) => `/user-vocabulary/${itemId}`,
		},
		DECKS: {
			LIST: "/vocabulary-decks",
			BY_ID: (id: ApiId) => `/vocabulary-decks/${id}`,
			ITEMS: "/vocabulary-deck-items",
			ITEM_BY_ID: (id: ApiId) => `/vocabulary-deck-items/${id}`,
		},
	},
	PERSONAS: {
		LIST: "/personas",
		BY_ID: (id: ApiId) => `/personas/${id}`,
	},
	SCENARIOS: {
		LIST: "/scenarios",
		BY_ID: (id: ApiId) => `/scenarios/${id}`,
		TOPICS: "/scenario-topics",
	},
	CONVERSATIONS: {
		LIST: "/conversations",
		BY_ID: (id: ApiId) => `/conversations/${id}`,
		CREATE: "/conversations",
		START: (id: ApiId) => `/conversations/${id}/start`,
		OFFER: (id: ApiId) => `/conversations/${id}/offer`,
		END: (id: ApiId) => `/conversations/${id}/end`,
		ANALYZE_TURN: (id: ApiId, position: number) =>
			`/conversations/${id}/turns/${position}/analyze`,
		// The correction modal. Go proxies transcript commands and retake audio
		// to the live engine session.
		TRANSCRIPT: (id: ApiId) => `/conversations/${id}/transcript`,
		TRANSCRIBE: (id: ApiId) => `/conversations/${id}/transcribe`,
	},
	LESSONS: {
		LIST: "/lessons",
		BY_ID: (id: ApiId) => `/lessons/${id}`,
		BY_SLUG: (slug: string) => `/lessons/by-slug/${slug}`,
		START_ATTEMPT: (id: ApiId) => `/lessons/${id}/attempts`,
		OPEN_ATTEMPT: (id: ApiId) => `/lessons/${id}/attempts/open`,
		CHECK: (activityId: ApiId) => `/lesson-activities/${activityId}/check`,
		SCORE: (activityId: ApiId) => `/lesson-activities/${activityId}/score`,
		PRONOUNCE: (activityId: ApiId) =>
			`/lesson-activities/${activityId}/pronounce`,
	},
	PROGRESS: {
		SUMMARY: "/progress",
	},
} as const;

/**
 * Server-side API base fallback for local dev (no Coolify env). The
 * production value comes from the `API_INTERNAL_URL` runtime env (bulk);
 * this literal lives here — the backend-URLs module, exempt from the
 * hardcoded-route guard — so services never spell an origin themselves.
 */
export const API_INTERNAL_URL_FALLBACK = "http://localhost:8000/api/v1";
