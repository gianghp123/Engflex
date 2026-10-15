// Shared API contracts. Generated files are produced by `tygo generate`
// in apps/api — do not edit them by hand. Shared hand-written helpers
// may live in this package alongside the barrel.
//
// tygo has no cross-package type resolution, so apps/api/tygo.yaml maps
// every `enums.*` type and `common.ListParams` to the bare names below.
// These ambient aliases make those names resolvable inside the generated
// files (which carry no imports — no frontmatter). They live here, not in a
// standalone globals.d.ts, because this barrel is the package entry point
// and is therefore always part of a consumer's TS program. Keep this list
// in sync with apps/api/internal/common/enums and common.ListParams.
import type * as Common from './common.ts';
import type * as Enums from './enums.ts';

declare global {
  type ActivityType = Enums.ActivityType;
  type AttemptType = Enums.AttemptType;
  type AttemptStatus = Enums.AttemptStatus;
  type CEFR = Enums.CEFR;
  type ConversationMode = Enums.ConversationMode;
  type ConversationStatus = Enums.ConversationStatus;
  type Goal = Enums.Goal;
  type Gender = Enums.Gender;
  type InterestTopic = Enums.InterestTopic;
  type LessonStatus = Enums.LessonStatus;
  type Level = Enums.Level;
  type ListParams = Common.ListParams;
  type ScenarioDifficulty = Enums.ScenarioDifficulty;
  type TurnRole = Enums.TurnRole;
  type SpanStatus = Enums.SpanStatus;
  type RelevanceStatus = Enums.RelevanceStatus;
  type VocabularySort = Enums.VocabularySort;
  type VocabularySource = Enums.VocabularySource;
  type VocabularyPartOfSpeech = Enums.VocabularyPartOfSpeech;
}

export * from './common.ts';
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
export * from './attempts-requests.ts';
export * from './attempts-responses.ts';
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
