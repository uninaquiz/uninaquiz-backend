package queries

type GetAllUsersQuery struct {
	Page  int `form:"page,default=1"  binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
