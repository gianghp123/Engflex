package common

// Wire-stable error codes for entitlement and quota failures. The web
// branches on these (surfaced through the envelope's data field), never on
// human message text. Values mirror the spec §9 contract and the OpenMeter
// reason taxonomy (spec §16.4).
const (
	// CodePremiumRequired marks a 403: the caller's plan does not include
	// the feature at all.
	CodePremiumRequired = "premium_required"
	// CodeFeatureNotInPlan marks a 403: the counter/feature exists but is
	// disabled for the plan.
	CodeFeatureNotInPlan = "feature_not_in_plan"
	// CodeInsufficientCredits marks a 429: the ai_credit limit is reached.
	CodeInsufficientCredits = "insufficient_credits"
	// CodeQuotaExceeded marks a 429: a non-credit counter (voice_session)
	// is reached.
	CodeQuotaExceeded = "quota_exceeded"
	// CodeRateLimited marks a 429: inline-AI per-user rate limit.
	CodeRateLimited = "rate_limited"
	// CodeTrialExpired marks a 403: the trial ended; free plan applies.
	CodeTrialExpired = "trial_expired"
	// CodePolicyUnavailable marks a 503: the policy/usage store is down;
	// paid AI fails closed (spec §12).
	CodePolicyUnavailable = "policy_unavailable"
	// CodeAIUnavailable marks a 503 retryable: the LLM/voice engine failed.
	// Completion-critical AI still attaches its fallback part.
	CodeAIUnavailable = "ai_unavailable"
)
