package dtos

type UpdateMineRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=1,max=50"`
}
