package create

type NewProductCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type GetProductCategoriesRequest struct {
	ProductCategories []string `json:"product_categories"`
}
