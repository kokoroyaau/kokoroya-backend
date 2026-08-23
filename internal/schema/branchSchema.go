package schema

type CreateBranchRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateBranchRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}
