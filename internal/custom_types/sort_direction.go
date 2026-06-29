package custom_types

import "strings"

type SortDirection string

const (
	Ascending  SortDirection = "ASC"
	Descending SortDirection = "DESC"
)

func (s SortDirection) IsValid() bool {
	cleaned := SortDirection(strings.ToUpper(strings.TrimSpace(string(s))))
	return cleaned == Ascending || cleaned == Descending
}

func (s SortDirection) Normalized() SortDirection {
	cleaned := strings.ToUpper(strings.TrimSpace(string(s)))
	if cleaned == "ASC" {
		return Ascending
	}
	return Descending
}
