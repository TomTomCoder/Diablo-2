package d2player

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
)

func TestInventorySlotItemCodeReadsLiveContainer(t *testing.T) {
	container := []string{"", d2hero.ItemPotionDeMana, ""}
	slot := &inventorySlot{container: &container, index: 1}

	if got := slot.itemCode(); got != d2hero.ItemPotionDeMana {
		t.Errorf("expected the current content at index 1, got %q", got)
	}

	// The whole point of reading through a pointer to the live container
	// (see newInventorySlot's own doc comment): a mutation after
	// construction must be visible on the next read, not just whatever
	// was there when the slot was built.
	container[1] = d2hero.ItemAnneauDuDebut

	if got := slot.itemCode(); got != d2hero.ItemAnneauDuDebut {
		t.Errorf("expected the updated content at index 1, got %q", got)
	}
}

func TestInventorySlotItemCodeOutOfRangeReturnsEmpty(t *testing.T) {
	container := []string{d2hero.ItemPotionDeMana}

	cases := []struct {
		name      string
		container *[]string
		index     int
	}{
		{"negative index", &container, -1},
		{"index past the end", &container, 5},
		{"nil container", nil, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			slot := &inventorySlot{container: c.container, index: c.index}

			if got := slot.itemCode(); got != "" {
				t.Errorf("expected empty string, got %q", got)
			}
		})
	}
}

func TestInventorySlotGetSetEnabledRoundTrips(t *testing.T) {
	slot := &inventorySlot{enabled: true}

	slot.SetEnabled(false)

	if slot.GetEnabled() {
		t.Fatal("expected GetEnabled to report false after SetEnabled(false)")
	}
}

func TestInventorySlotGetSetPressedRoundTrips(t *testing.T) {
	slot := &inventorySlot{}

	slot.SetPressed(true)

	if !slot.GetPressed() {
		t.Fatal("expected GetPressed to report true after SetPressed(true)")
	}
}

func TestInventorySlotActivateRunsTheRegisteredCallback(t *testing.T) {
	slot := &inventorySlot{}
	ran := false

	slot.OnActivated(func() { ran = true })
	slot.Activate()

	if !ran {
		t.Fatal("expected Activate to run the callback registered via OnActivated")
	}
}

func TestInventorySlotActivateWithNoCallbackDoesNotPanic(t *testing.T) {
	slot := &inventorySlot{}

	slot.Activate() // must not panic: this is the state every slot starts in
}
