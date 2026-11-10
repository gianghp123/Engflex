package requests

type PatchAttempt struct {
	Revision int      `json:"revision" binding:"required,min=1" example:"1"`
	Answers  []Answer `json:"answers" binding:"max=200,dive"`
	Submit   bool     `json:"submit" example:"false"`
}
type Answer struct {
	QuestionID   string   `json:"questionId" binding:"required"`
	SelectedKey  *string  `json:"selectedKey,omitempty" tstype:"string | null"`
	SelectedKeys []string `json:"selectedKeys,omitempty" tstype:"string[] | null"`
	TextValue    *string  `json:"textValue,omitempty" tstype:"string | null"`
	Flagged      bool     `json:"flagged"`
}
