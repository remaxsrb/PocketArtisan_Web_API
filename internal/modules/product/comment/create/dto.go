package create

type Request struct {
	ProductID  uint64   `json:"productID" binding:"required"`
	Text       string   `json:"text" binding:"required"`
	PhotoLinks []string `json:"photoLinks"`
	VideoLinks []string `json:"videoLinks"`
}
