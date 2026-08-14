package d2player

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2geom"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

// newTestSkillPanel builds a single-row, single-skill panel with its row
// positioned exactly where the real getRowStartX/generateSkillRowImageCache
// would place it -- getSkillAtPos derives which skill was clicked from that
// same start position, so the test click coordinates below (rowStartX, 0)
// must agree with it or every click resolves to the wrong (or an
// out-of-range) skill index.
func newTestSkillPanel(t *testing.T, isLeftPanel bool, skillID int) (panel *SkillPanel, rowStartX int) {
	t.Helper()

	skill := &d2hero.HeroSkill{SkillRecord: &d2records.SkillRecord{ID: skillID}}
	width := skillIconWidth * 1

	rowStartX = leftPanelStartX
	if !isLeftPanel {
		rowStartX = rightPanelEndX - width
	}

	row := &SkillListRow{
		Skills:    []*d2hero.HeroSkill{skill},
		Rectangle: d2geom.Rectangle{Left: rowStartX, Top: 0, Width: width, Height: skillIconHeight},
	}

	panel = &SkillPanel{
		isOpen:      true,
		isLeftPanel: isLeftPanel,
		ListRows:    []*SkillListRow{row},
	}

	return panel, rowStartX
}

func TestSkillPanelHandleClickRequestsTheLeftSlotFromTheLeftPanel(t *testing.T) {
	s, x := newTestSkillPanel(t, true, 42)

	var gotSlot d2hero.SkillSlot

	var gotSkillID int

	s.SetOnEquipCb(func(slot d2hero.SkillSlot, skillID int) {
		gotSlot, gotSkillID = slot, skillID
	})

	if !s.HandleClick(x, 0) {
		t.Fatal("expected HandleClick to report the click as handled")
	}

	if gotSlot != d2hero.SkillSlotLeft || gotSkillID != 42 {
		t.Errorf("expected (SkillSlotLeft, 42), got (%v, %d)", gotSlot, gotSkillID)
	}
}

func TestSkillPanelHandleClickRequestsTheRightSlotFromTheRightPanel(t *testing.T) {
	s, x := newTestSkillPanel(t, false, 7)

	var gotSlot d2hero.SkillSlot

	s.SetOnEquipCb(func(slot d2hero.SkillSlot, _ int) {
		gotSlot = slot
	})

	s.HandleClick(x, 0)

	if gotSlot != d2hero.SkillSlotRight {
		t.Errorf("expected SkillSlotRight, got %v", gotSlot)
	}
}

func TestSkillPanelHandleClickOutsideThePanelDoesNotFireTheCallback(t *testing.T) {
	s, _ := newTestSkillPanel(t, true, 42)

	fired := false

	s.SetOnEquipCb(func(d2hero.SkillSlot, int) { fired = true })

	if s.HandleClick(9999, 9999) {
		t.Error("expected HandleClick to report the click as unhandled outside the panel")
	}

	if fired {
		t.Error("expected the callback not to fire for a click outside the panel")
	}
}

func TestSkillPanelHandleClickWithNoCallbackDoesNotPanic(t *testing.T) {
	s, x := newTestSkillPanel(t, true, 42)

	if !s.HandleClick(x, 0) {
		t.Error("expected HandleClick to still report the click as handled")
	}
}
