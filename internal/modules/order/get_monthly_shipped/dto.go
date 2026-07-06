package get_monthly_shipped

type MonthlyShippedRequest struct {
	From string `form:"from" query:"from"`
	To   string `form:"to" query:"to"`
}

type MonthlyShippedCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}