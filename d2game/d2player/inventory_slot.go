package d2player

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

const (
	inventorySlotWidth  = 90
	inventorySlotHeight = 16
	inventorySlotEmpty  = "-"
)

// newInventorySlot returns a clickable widget for one fixed slot index into
// container (HeroState.Inventory/Stash/Belt's client-side mirror,
// d2mapentity.Player.Inventory/Stash -- see HeroDevil's own architecture
// notes on why d2ui has no widget-removal capability, which is why this
// exists at all): container is a pointer to the live slice, read fresh
// every Render() call rather than captured once, so the slot always shows
// current content even though the widget itself is built once and never
// rebuilt. Devil has no item icon art of any kind (ROADMAP.md Phase 6),
// so this renders as plain text via Surface.DrawTextf -- the same asset-
// independent primitive DevilHUD already relies on -- rather than
// inventing an icon-sheet mapping the way skillIcon's IconCel reuse does.
func newInventorySlot(ui *d2ui.UIManager, container *[]string, index int) *inventorySlot {
	base := d2ui.NewBaseWidget(ui)

	res := &inventorySlot{
		BaseWidget: base,
		container:  container,
		index:      index,
		enabled:    true,
	}

	res.SetSize(inventorySlotWidth, inventorySlotHeight)

	return res
}

type inventorySlot struct {
	*d2ui.BaseWidget
	container  *[]string
	index      int
	enabled    bool
	pressed    bool
	onActivate func()
}

// static check: WidgetGroup.AddWidget only registers a widget as
// clickable via a runtime type assertion (w.(d2ui.ClickableWidget)) --
// same precedent as skillIcon's own identical guard.
var _ d2ui.ClickableWidget = &inventorySlot{}

// itemCode returns the item code currently in this slot ("" if empty),
// bounds-checked against the live container -- container's length is
// fixed once HeroState.InitStorage/InitBelt runs, but this guards against
// ever reading past it regardless (same defensive practice as the
// handleItemMovedTo*/From* packet handlers).
func (s *inventorySlot) itemCode() string {
	if s.container == nil || s.index < 0 || s.index >= len(*s.container) {
		return ""
	}

	return (*s.container)[s.index]
}

// GetEnabled returns whether the slot currently accepts clicks.
func (s *inventorySlot) GetEnabled() bool { return s.enabled }

// SetEnabled sets whether the slot currently accepts clicks.
func (s *inventorySlot) SetEnabled(enabled bool) { s.enabled = enabled }

// GetPressed returns whether the slot is currently held down.
func (s *inventorySlot) GetPressed() bool { return s.pressed }

// SetPressed sets whether the slot is currently held down.
func (s *inventorySlot) SetPressed(pressed bool) { s.pressed = pressed }

// OnActivated registers the callback run when the slot is clicked. The
// server is the final authority on whether the move is actually valid
// (empty source, full destination, non-potion into the belt...) -- same
// "let the server reject it, don't try to predict that client-side"
// practice already used for LearnSkillPreview's not-yet-eligible skills.
func (s *inventorySlot) OnActivated(callback func()) { s.onActivate = callback }

// Activate runs the click callback, if one was registered.
func (s *inventorySlot) Activate() {
	if s.onActivate != nil {
		s.onActivate()
	}
}

// Render draws this slot's current item code (or inventorySlotEmpty) as
// plain text at its position.
func (s *inventorySlot) Render(target d2interface.Surface) {
	x, y := s.GetPosition()

	text := s.itemCode()
	if text == "" {
		text = inventorySlotEmpty
	}

	target.PushTranslation(x, y)
	defer target.Pop()

	target.DrawTextf("%s", text)
}

func (s *inventorySlot) Advance(elapsed float64) error {
	return nil
}
