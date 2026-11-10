package enums

// ExamSkill identifies the specific language skill tested (exam_tests.skill).
type ExamSkill string

const (
	ExamSkillReading   ExamSkill = "reading"
	ExamSkillListening ExamSkill = "listening"
	ExamSkillWriting   ExamSkill = "writing"
	ExamSkillSpeaking  ExamSkill = "speaking"
	ExamSkillFull      ExamSkill = "full"
)
