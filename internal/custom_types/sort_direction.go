package custom_types

import "strings"

type SortDirection string

const (
	Ascending  SortDirection = "ASC"
	Descending SortDirection = "DESC"
)

func (s SortDirection) IsValid() bool {
	switch SortDirection(strings.ToUpper(string(s))) {
	case Ascending, Descending:
		return true
	default:
		return false
	}
}
