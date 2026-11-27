type ApiId = string | number;

/** Backend paths only. No hardcoded URLs in services — import from here. */
export const API_ROUTES = {
	VIDEO_TRANSCRIPTS: {
		LIST: "/video-transcripts",
		BY_ID: (id: ApiId) => `/video-transcripts/${id}`,
		BY_VIDEO: (id: ApiId) => `/video-transcripts/video-exercise/${id}`,
	},
	PROFILES: {
		ME: "/profiles/me",
		BY_ID: (id: ApiId) => `/profiles/${id}`,
	},
	PRONUNCIATION: {
		RESPEAK: "/pronunciation/respeak",
	},
	VOCABULARY: {
		LIST: "/vocabulary",
		BY_ID: (id: ApiId) => `/vocabulary/${id}`,
	},
	VOCABULARY_ITEMS: {
		LIST: "/vocabulary-items",
		BY_ID: (id: ApiId) => `/vocabulary-items/${id}`,
	},
	VOCABULARY_DECKS: {
		LIST: "/vocabulary-decks",
		BY_ID: (id: ApiId) => `/vocabulary-decks/${id}`,
	},
	VOCABULARY_DECK_ITEMS: {
		LIST: "/vocabulary-deck-items",
		BY_ID: (id: ApiId) => `/vocabulary-deck-items/${id}`,
	},
	VOCABULARY_CATEGORIES: {
		LIST: "/vocabulary-categories",
		BY_ID: (id: ApiId) => `/vocabulary-categories/${id}`,
	},
	USER_VOCABULARY: {
		LIST: "/user-vocabulary",
		BY_ITEM_ID: (itemId: ApiId) => `/user-vocabulary/${itemId}`,
		NOTE: (itemId: ApiId) => `/user-vocabulary/${itemId}/note`,
	},
	VIDEO_EXERCISES: {
		LIST: "/video-exercises",
		BY_ID: (id: ApiId) => `/video-exercises/${id}`,
	},
	VIDEO_CATEGORIES: {
		LIST: "/video-categories",
		BY_ID: (id: ApiId) => `/video-categories/${id}`,
	},
	LESSON_SECTIONS: {
		LIST: "/lesson-sections",
		BY_ID: (id: ApiId) => `/lesson-sections/${id}`,
	},
	LESSON_ACTIVITIES: {
		LIST: "/lesson-activities",
		BY_ID: (id: ApiId) => `/lesson-activities/${id}`,
	},
	ADMIN: {
		VOCABULARY_DECKS: "/admin/vocabulary-decks",
		VOCABULARY_ITEMS: "/admin/vocabulary-items",
		LESSON_ACTIVITY: (id: ApiId) => `/admin/lesson-activities/${id}`,
		USERS: "/admin/users",
		USER_ROLE: (id: ApiId) => `/admin/users/${id}/role`,
		USER_SUBSCRIPTIONS: (id: ApiId) => `/admin/users/${id}/subscriptions`,
		PREMIUM_PLANS: "/admin/premium/plans",
		PREMIUM_PLAN: (plan: ApiId) => `/admin/premium/plans/${plan}`,
		PREMIUM_FEATURES: "/admin/premium/features",
		PREMIUM_FEATURE: (id: ApiId) => `/admin/premium/features/${id}`,
		PREMIUM_SUBSCRIPTIONS: "/admin/premium/subscriptions",
		PREMIUM_LIMITS: "/admin/premium/limits",
		PREMIUM_LIMIT: (plan: ApiId) => `/admin/premium/limits/${plan}`,
		USER_USAGE: (id: ApiId) => `/admin/users/${id}/usage`,
		USER_TRIAL: (id: ApiId) => `/admin/users/${id}/trial`,
	},
	PAYMENTS: {
		PLANS: "/payments/vnpay/plans",
		CHECKOUT: "/payments/vnpay/checkout",
		ME: "/premium/me",
		TRIAL: "/premium/trial",
		USAGE: "/premium/usage",
		SEPAY_CHECKOUT: "/payments/sepay/checkout",
		SEPAY_STATUS: (reference: string) => `/payments/sepay/status/${reference}`,
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
		SUMMARIZE: (id: ApiId) => `/conversations/${id}/summarize`,
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
		REDACT_WRITING: (attemptId: ApiId) =>
			`/lesson-activities/attempts/${attemptId}/writing`,
	},
	PROGRESS: {
		SUMMARY: "/progress",
		ERRORS: "/progress/errors",
		MISTAKES: "/progress/mistakes",
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
			PRONOUNCE: (id: ApiId, attemptId: ApiId) =>
				`/video-exercises/${id}/attempts/${attemptId}/pronounce`,
		},
	},
	STORIES: {
		LIST: "/stories",
		BY_ID: (id: ApiId) => `/stories/${id}`,
		BY_SLUG: (slug: string) => `/stories/by-slug/${slug}`,
		TOPICS: "/stories/topics",
		UPLOAD: "/stories/upload",
	},
	BOOKMARKS: {
		LIST: "/bookmarks",
		LIBRARY: "/bookmarks/library",
		CREATE: "/bookmarks",
		DELETE: (contentType: string, contentId: ApiId) =>
			`/bookmarks/${contentType}/${contentId}`,
		NOTE: (contentType: string, contentId: ApiId) =>
			`/bookmarks/${contentType}/${contentId}/note`,
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
			STATISTICS: "/exam-attempts/statistics",
			CANCEL: (id: ApiId) => `/exam-attempts/${id}/cancel`,
			BY_ID: (id: ApiId) => `/exam-attempts/${id}`,
			AI_REVIEW: (id: ApiId) => `/exam-attempts/${id}/ai-review`,
			STUDY_PLAN: (id: ApiId) => `/exam-attempts/${id}/study-plan`,
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
