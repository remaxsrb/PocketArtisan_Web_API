package craftsman_application

import (
	"PocketArtisan/internal/entities"
	"testing"
)

func TestNextApplicationStatus_AllowsValidTransitions(t *testing.T) {
	approved, err := NextApplicationStatus(entities.StatusPending, ApplicationActionApprove)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved != entities.StatusAccepted {
		t.Fatalf("want %s, got %s", entities.StatusAccepted, approved)
	}

	rejected, err := NextApplicationStatus(entities.StatusPending, ApplicationActionReject)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected != entities.StatusRejected {
		t.Fatalf("want %s, got %s", entities.StatusRejected, rejected)
	}
}

func TestNextApplicationStatus_RejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		name   string
		status string
		action ApplicationAction
	}{
		{name: "approved cannot reject", status: entities.StatusAccepted, action: ApplicationActionReject},
		{name: "rejected cannot approve", status: entities.StatusRejected, action: ApplicationActionApprove},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NextApplicationStatus(tc.status, tc.action)
			if err == nil {
				t.Fatal("expected transition error")
			}
		})
	}
}

func TestInitialApplicationStatus(t *testing.T) {
	if got := InitialApplicationStatus(); got != entities.StatusPending {
		t.Fatalf("want %s, got %s", entities.StatusPending, got)
	}
}
