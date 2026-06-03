package sort

type Property string

const (
	Craft Property = "craft"
)

type SortRequest struct {
	Property Property `json:"direction"` // next | prev
}
