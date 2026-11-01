type ApiId = string | number;

/** Backend paths only. No hardcoded URLs in services — import from here. */
export const API_ROUTES = {
	PROFILES: {
		ME: "/profiles/me",
		BY_ID: (id: ApiId) => `/profiles/${id}`,
	},
	VOCABULARY: {
		ITEMS: {
			LIST: "/vocabulary-items",
			BY_ID: (id: ApiId) => `/vocabulary-items/${id}`,
			GENERATE: "/vocabulary-items/generate",
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
		PROGRESS: "/lessons/progress",
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
	VIDEOS: {
		LIST: "/video-exercises",
		BY_ID: (id: ApiId) => `/video-exercises/${id}`,
		CATEGORIES_PREVIEW: "/video-categories/preview",
		ATTEMPTS: {
			START: (id: ApiId) => `/video-exercises/${id}/attempts`,
			OPEN: (id: ApiId) => `/video-exercises/${id}/attempts/open`,
			SAVE_PROGRESS: (id: ApiId, attemptId: ApiId) =>
				`/video-exercises/${id}/attempts/${attemptId}/progress`,
			COMPLETE: (id: ApiId, attemptId: ApiId) =>
				`/video-exercises/${id}/attempts/${attemptId}/complete`,
		},
	},
	STORIES: {
		LIST: "/stories",
		BY_ID: (id: ApiId) => `/stories/${id}`,
		BY_SLUG: (slug: string) => `/stories/by-slug/${slug}`,
		TOPICS: "/stories/topics",
	},
	BOOKMARKS: {
		LIST: "/bookmarks",
		CREATE: "/bookmarks",
		DELETE: (contentType: string, contentId: ApiId) =>
			`/bookmarks/${contentType}/${contentId}`,
	},
	EXAMS: {
		CATEGORIES: {
			LIST: "/exam-categories",
			TESTS: (categoryId: ApiId) => `/exam-categories/${categoryId}/tests`,
		},
		TESTS: {
			BY_ID: (id: ApiId) => `/exam-tests/${id}`,
			START_ATTEMPT: (id: ApiId) => `/exam-tests/${id}/attempts`,
		},
		ATTEMPTS: {
			LIST: "/exam-attempts",
			STATS: "/exam-attempts/stats",
			CANCEL: (id: ApiId) => `/exam-attempts/${id}/cancel`,
			BY_ID: (id: ApiId) => `/exam-attempts/${id}`,
		},
	},
} as const;

/**
 * Server-side API base fallback for local dev (no Coolify env). The
 * production value comes from the `API_INTERNAL_URL` runtime env (bulk);
 * this literal lives here — the backend-URLs module, exempt from the
 * hardcoded-route guard — so services never spell an origin themselves.
 */
export const API_INTERNAL_URL_FALLBACK = "http://localhost:8000/api/v1";
