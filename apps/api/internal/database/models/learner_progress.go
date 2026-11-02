package models

// LearnerProgress is an aggregate read model, never a persisted table.
type LearnerProgress struct {
	StreakDays        int
	VoiceMinutesToday int
	DailyGoalMinutes  int
	LessonsCompleted  int
}
