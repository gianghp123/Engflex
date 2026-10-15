package enums

// Goal is one onboarding goal (user_profiles.preferences.goals[]).
type Goal string

const (
	GoalTravel Goal = "travel"
	GoalWork   Goal = "work"
	GoalExams  Goal = "exams"
	GoalDaily  Goal = "daily"
	GoalFun    Goal = "fun"
)
