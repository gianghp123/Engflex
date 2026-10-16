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
} as const;
