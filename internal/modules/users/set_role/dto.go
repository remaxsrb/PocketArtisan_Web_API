package set_role

type SetRoleRequest struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required"`
}
