package d2player

import "testing"

func TestSkillIconGetSetEnabledRoundTrips(t *testing.T) {
	si := &skillIcon{enabled: true}

	si.SetEnabled(false)

	if si.GetEnabled() {
		t.Fatal("expected GetEnabled to report false after SetEnabled(false)")
	}
}

func TestSkillIconGetSetPressedRoundTrips(t *testing.T) {
	si := &skillIcon{}

	si.SetPressed(true)

	if !si.GetPressed() {
		t.Fatal("expected GetPressed to report true after SetPressed(true)")
	}
}

func TestSkillIconActivateRunsTheRegisteredCallback(t *testing.T) {
	si := &skillIcon{}
	ran := false

	si.OnActivated(func() { ran = true })
	si.Activate()

	if !ran {
		t.Fatal("expected Activate to run the callback registered via OnActivated")
	}
}

func TestSkillIconActivateWithNoCallbackDoesNotPanic(t *testing.T) {
	si := &skillIcon{}

	si.Activate() // must not panic: this is the state every skillIcon starts in
}
