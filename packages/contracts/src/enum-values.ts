// Code generated from ./enums.ts by scripts/gen-enum-values.mjs. DO NOT EDIT.
import type * as Enums from "./enums.ts";

export const ACTIVITY_TYPE_VALUES = ["reading", "listening", "writing", "speaking"] as const satisfies readonly Enums.ActivityType[];
export const ATTEMPT_STATUS_VALUES = ["in_progress", "completed", "abandoned"] as const satisfies readonly Enums.AttemptStatus[];
export const ATTEMPT_TYPE_VALUES = ["reading", "listening", "writing", "speaking", "shadowing", "flashcard", "benchmark", "dictation", "lesson", "exam"] as const satisfies readonly Enums.AttemptType[];
export const CEFR_VALUES = ["A1", "A2", "B1", "B2", "C1", "C2"] as const satisfies readonly Enums.CEFR[];
export const CONTENT_TYPE_VALUES = ["lesson", "video", "story"] as const satisfies readonly Enums.ContentType[];
export const CONVERSATION_MODE_VALUES = ["free_talk", "roleplay"] as const satisfies readonly Enums.ConversationMode[];
export const CONVERSATION_STATUS_VALUES = ["pending", "live", "ended", "failed"] as const satisfies readonly Enums.ConversationStatus[];
export const EXAM_MODE_VALUES = ["exam", "practice"] as const satisfies readonly Enums.ExamMode[];
export const GENDER_VALUES = ["male", "female"] as const satisfies readonly Enums.Gender[];
export const GOAL_VALUES = ["travel", "work", "exams", "daily", "fun"] as const satisfies readonly Enums.Goal[];
export const INTEREST_TOPIC_VALUES = ["travel", "business", "movies_tv", "technology", "food", "music", "sports", "news"] as const satisfies readonly Enums.InterestTopic[];
export const LESSON_STATUS_VALUES = ["unstarted", "in_progress", "completed"] as const satisfies readonly Enums.LessonStatus[];
export const LEVEL_VALUES = ["beginner", "intermediate", "advanced"] as const satisfies readonly Enums.Level[];
export const PAYMENT_TRANSACTION_STATUS_VALUES = ["pending", "succeeded", "failed"] as const satisfies readonly Enums.PaymentTransactionStatus[];
export const PUBLISH_STATUS_VALUES = ["draft", "published"] as const satisfies readonly Enums.PublishStatus[];
export const SCENARIO_DIFFICULTY_VALUES = ["A1", "A2", "B1", "B1+", "B2", "C1"] as const satisfies readonly Enums.ScenarioDifficulty[];
export const SPAN_STATUS_VALUES = ["incorrect", "awkward"] as const satisfies readonly Enums.SpanStatus[];
export const SUBSCRIPTION_PLAN_VALUES = ["1m", "3m", "12m"] as const satisfies readonly Enums.SubscriptionPlan[];
export const SUBSCRIPTION_STATUS_VALUES = ["pending", "active", "expired", "cancelled", "failed"] as const satisfies readonly Enums.SubscriptionStatus[];
export const TURN_ROLE_VALUES = ["user", "ai"] as const satisfies readonly Enums.TurnRole[];
export const VOCABULARY_DOMAIN_VALUES = ["backend_db", "distributed_systems", "devops_cloud", "frontend_ui", "ai_ml"] as const satisfies readonly Enums.VocabularyDomain[];
export const VOCABULARY_PART_OF_SPEECH_VALUES = ["noun", "verb", "adjective", "adverb", "pronoun", "preposition", "conjunction", "interjection", "determiner", "phrase", "idiom", "phrasal_verb"] as const satisfies readonly Enums.VocabularyPartOfSpeech[];
export const VOCABULARY_SORT_VALUES = ["recent", "mastery", "alphabetical", "interval"] as const satisfies readonly Enums.VocabularySort[];
export const VOCABULARY_SOURCE_VALUES = ["lesson", "conversation", "video", "manual", "story"] as const satisfies readonly Enums.VocabularySource[];
