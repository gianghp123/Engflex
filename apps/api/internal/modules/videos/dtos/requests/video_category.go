package requests

type CreateVideoCategory struct {
	Slug string `json:"slug" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdateVideoCategory struct {
	Slug string `json:"slug" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type PreviewVideoCategories struct {
	PreviewK int `form:"previewK" json:"previewK" binding:"required,min=1,max=32"`
}
