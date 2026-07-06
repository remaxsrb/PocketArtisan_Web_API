package get_monthly_shipped_by_category

type MonthlyShippedByCategoryRequest struct {
	CraftsmanID string `form:"craftsman_id" query:"craftsman_id"`
	From        string `form:"from" query:"from"`
	To          string `form:"to" query:"to"`
}

type MonthlyCategoryCount struct {
	Month    string `json:"month"`
	Category string `json:"category"`
	Count    int64  `json:"count"`
}