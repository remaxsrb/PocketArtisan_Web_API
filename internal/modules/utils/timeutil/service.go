package timeutil

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

type Service interface {
	// ParseDateRange parses optional "from"/"to" date strings (YYYY-MM-DD) into
	// a half-open [from, to) time range. The "to" bound is advanced to the
	// start of the next day so the supplied date is fully included, rather
	// than cut off at its midnight instant.
	ParseDateRange(from, to string) (*time.Time, *time.Time, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ParseDateRange(from, to string) (*time.Time, *time.Time, error) {
	var fromT, toT *time.Time

	if from != "" {
		t, err := time.Parse(DateLayout, from)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid from date: %w", err)
		}
		fromT = &t
	}

	if to != "" {
		t, err := time.Parse(DateLayout, to)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid to date: %w", err)
		}
		t = t.AddDate(0, 0, 1)
		toT = &t
	}

	return fromT, toT, nil
}