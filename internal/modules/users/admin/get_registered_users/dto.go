package get_registered_users

type Request struct {
	From string `form:"from" query:"from"`
	To   string `form:"to" query:"to"`
}

type Response struct {
	Total int64 `json:"total"`
}