package get_registered_users_by_month

type Request struct {
	From string `form:"from" query:"from"`
	To   string `form:"to" query:"to"`
}

type Bucket struct {
	Month string `json:"month"` // "2006-01"
	Total int64  `json:"total"`
}

type Response struct {
	Buckets []Bucket `json:"buckets"`
}
