package d2server

import (
	"testing"
	"time"
)

func TestEventIsInactiveByDefault(t *testing.T) {
	server := serverWithConnection(nil)

	if server.IsEventActive(EventTempeteDeLoot) {
		t.Error("expected an event that was never activated to be inactive")
	}
}

func TestActivateEventMakesItActiveUntilItExpires(t *testing.T) {
	server := serverWithConnection(nil)

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.ActivateEvent(EventTempeteDeLoot, time.Hour)

	if !server.IsEventActive(EventTempeteDeLoot) {
		t.Error("expected the event to be active immediately after ActivateEvent")
	}

	server.clock = func() time.Time { return now.Add(2 * time.Hour) }

	if server.IsEventActive(EventTempeteDeLoot) {
		t.Error("expected the event to be inactive after its duration elapsed")
	}
}

func TestActivateEventsAreIndependent(t *testing.T) {
	server := serverWithConnection(nil)

	server.ActivateEvent(EventTempeteDeLoot, time.Hour)

	if server.IsEventActive(EventNuitDeLApocalypse) {
		t.Error("expected a different, never-activated event to remain inactive")
	}
}

func TestReactivatingAnEventResetsItsExpiry(t *testing.T) {
	server := serverWithConnection(nil)

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.ActivateEvent(EventTempeteDeLoot, time.Hour)

	server.clock = func() time.Time { return now.Add(59 * time.Minute) }
	server.ActivateEvent(EventTempeteDeLoot, time.Hour) // reactivate just before expiry

	server.clock = func() time.Time { return now.Add(90 * time.Minute) }

	if !server.IsEventActive(EventTempeteDeLoot) {
		t.Error("expected reactivation to reset the expiry from the reactivation time, not stack")
	}
}
