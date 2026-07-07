package craftsman_application

import (
	"PocketArtisan/internal/entities"
	"fmt"
)

type ApplicationAction string

const (
	ApplicationActionApprove ApplicationAction = "approve"
	ApplicationActionReject  ApplicationAction = "reject"
)

type applicationState interface {
	status() string
	transition(action ApplicationAction) (string, error)
}

type pendingApplicationState struct{}
type approvedApplicationState struct{}
type rejectedApplicationState struct{}

func (pendingApplicationState) status() string  { return entities.StatusPending }
func (approvedApplicationState) status() string { return entities.StatusAccepted }
func (rejectedApplicationState) status() string { return entities.StatusRejected }

func (pendingApplicationState) transition(action ApplicationAction) (string, error) {
	switch action {
	case ApplicationActionApprove:
		return entities.StatusAccepted, nil
	case ApplicationActionReject:
		return entities.StatusRejected, nil
	default:
		return "", invalidApplicationTransition(entities.StatusPending, action)
	}
}

func (approvedApplicationState) transition(action ApplicationAction) (string, error) {
	return "", invalidApplicationTransition(entities.StatusAccepted, action)
}

func (rejectedApplicationState) transition(action ApplicationAction) (string, error) {
	return "", invalidApplicationTransition(entities.StatusRejected, action)
}

func NextApplicationStatus(current string, action ApplicationAction) (string, error) {
	state, err := resolveApplicationState(current)
	if err != nil {
		return "", err
	}
	return state.transition(action)
}

func InitialApplicationStatus() string {
	return entities.StatusPending
}

func resolveApplicationState(status string) (applicationState, error) {
	switch status {
	case entities.StatusPending:
		return pendingApplicationState{}, nil
	case entities.StatusAccepted:
		return approvedApplicationState{}, nil
	case entities.StatusRejected:
		return rejectedApplicationState{}, nil
	default:
		return nil, fmt.Errorf("unknown application status: %s", status)
	}
}

func invalidApplicationTransition(current string, action ApplicationAction) error {
	return fmt.Errorf("invalid application transition: status=%s action=%s", current, action)
}
