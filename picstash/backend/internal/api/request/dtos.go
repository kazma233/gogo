package request

type SendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type UpdateTagsRequest struct {
	TagIDs []int `json:"tag_ids" validate:"required"`
}

type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Color string `json:"color" validate:"required,len=7"`
}

type UpdateTagRequest struct {
	Name  string `json:"name" validate:"omitempty,min=1,max=50"`
	Color string `json:"color" validate:"omitempty,len=7"`
}

type UploadRequest struct {
	Files  []string `json:"files" validate:"required"`
	TagIDs []int    `json:"tag_ids"`
}
