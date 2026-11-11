// Shared API contracts. Generated files are produced by `tygo generate`
// in apps/api — do not edit them by hand. Shared hand-written helpers
// may live in this package alongside the barrel.
//
// tygo has no cross-package type resolution, so apps/api/tygo.yaml maps
// every `enums.*` type, `common.ListParams` and `llm.LanguageParam` to the
// bare names below. These ambient aliases make those names resolvable inside
// the generated files (which carry no imports — no frontmatter). They live
// here, not in a standalone globals.d.ts, because this barrel is the package
// entry point and is therefore always part of a consumer's TS program. Keep
// this list in sync with apps/api/internal/common/enums, common.ListParams
// and internal/llm.LanguageParam.
import type * as Common from './common.ts';
import type * as Enums from './enums.ts';
import type * as Llm from './llm.ts';

declare global {
  type ActivityType = Enums.ActivityType;
  type AttemptType = Enums.AttemptType;
  type AttemptStatus = Enums.AttemptStatus;
  type ExamMode = Enums.ExamMode;
  type ExamType = Enums.ExamType;
  type ExamSkill = Enums.ExamSkill;
  type ExamSubType = Enums.ExamSubType;
  type CEFR = Enums.CEFR;
  type ConversationMode = Enums.ConversationMode;
  type ConversationStatus = Enums.ConversationStatus;
  type Goal = Enums.Goal;
  type Gender = Enums.Gender;
  type InterestTopic = Enums.InterestTopic;
  type LessonStatus = Enums.LessonStatus;
  type Level = Enums.Level;
  type ListParams = Common.ListParams;
  type PublishStatus = Enums.PublishStatus;
  type LanguageParam = Llm.LanguageParam;
  type ScenarioDifficulty = Enums.ScenarioDifficulty;
  type TurnRole = Enums.TurnRole;
  type SpanStatus = Enums.SpanStatus;
  type ErrorType = Enums.ErrorType;
  type RelevanceStatus = Enums.RelevanceStatus;
  type VocabularySort = Enums.VocabularySort;
  type VocabularySource = Enums.VocabularySource;
  type VocabularyDomain = Enums.VocabularyDomain;
  type SubscriptionStatus = Enums.SubscriptionStatus;
  type SubscriptionPlan = Enums.SubscriptionPlan;
  type PaymentTransactionStatus = Enums.PaymentTransactionStatus;
  type VocabularyPartOfSpeech = Enums.VocabularyPartOfSpeech;
  type ContentType = Enums.ContentType;
}

export * from './common.ts';
export * from './llm.ts';
export * from './enums.ts';
export * from './enum-values.ts';
export * from './profiles-requests.ts';
export * from './profiles-responses.ts';
export * from './lessons-requests.ts';
export * from './lessons-responses.ts';
export * from './scenarios-requests.ts';
export * from './scenarios-responses.ts';
export * from './conversations-requests.ts';
export * from './conversations-responses.ts';
export * from './vocabulary-requests.ts';
export * from './vocabulary-responses.ts';
export type {
  PatchAttempt as PatchExamAttempt,
  StartAttempt as StartExamAttempt,
  Answer as ExamAnswer,
} from './exams-requests.ts';

export type {
  Category as ExamCategory,
  TestMetadata as ExamTestMetadata,
  TestDetail as ExamTestDetail,
  ExamAttempt,
  ExamAttemptSummary,
  ExamStatistics,
  ReviewAttempt as ExamReviewAttempt,
  TakingContent as ExamTakingContent,
  ReviewContent as ExamReviewContent,
  IeltsContentMetadata,
  IeltsSectionMetadata,
  IeltsTakingContent,
  IeltsTakingSection,
  IeltsPassage,
  IeltsParagraph,
  IeltsTakingQuestionGroup,
  IeltsTakingQuestion,
  IeltsReviewContent,
  IeltsReviewSection,
  IeltsReviewQuestionGroup,
  IeltsReviewQuestion,
} from './exams-responses.ts';
export * from './videos-requests.ts';
export * from './videos-responses.ts';
export * from './attempts-requests.ts';
export * from './attempts-responses.ts';
export * from './users-requests.ts';
export * from './users-responses.ts';
export * from './payments-requests.ts';
export * from './payments-responses.ts';
export * from './bookmarks-requests.ts';
export * from './bookmarks-responses.ts';
export * from './stories-requests.ts';
export * from './stories-responses.ts';
// Both conversations and lessons responses declare structurally identical
// MispronouncedWord/PhoneDetail/Prosody/ReferencePhone copies; an explicit
// re-export resolves the star-export ambiguity (TS2308). Structural typing
// keeps both wire copies compatible.
export type {
  MispronouncedWord,
  PhoneDetail,
  Prosody,
  ReferencePhone,
} from './conversations-responses.ts';
// Both vocabulary requests and responses declare structurally identical
// VocabularyDetails/VocabularyContext/Collocation/WordForm copies (the
// request side is a wire copy because tygo cannot resolve cross-package
// types); an explicit re-export resolves the star-export ambiguity (TS2308).
// Structural typing keeps both wire copies compatible.
export type {
  VocabularyDetails,
  VocabularyContext,
  Collocation,
  WordForm,
} from './vocabulary-responses.ts';
