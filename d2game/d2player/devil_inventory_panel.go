package d2player

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

// devilInventoryPanel is a real, clickable screen for HeroState.
// Inventory/Stash's already-tested MoveToStash/MoveToInventory mechanism
// and network round trip -- neither of which had any UI to trigger them
// (ROADMAP.md Phase 5) until now. Deliberately not built on top of the
// pre-existing Inventory/ItemGrid (inventory.go): that panel renders stock
// Diablo 2's own spatial item-grid data model (real InventoryItem objects
// with width/height), a different, incompatible shape from Devil's own
// flat []string slot lists -- see ROADMAP.md's own architecture notes on
// why. A brand new screen instead, showing both containers as fixed 8x5
// grids (Devil's Inventory/Stash are both a flat 40 slots,
// d2hero.InitStorage) of plain-text slots (inventorySlot) -- clicking an
// Inventory slot requests MoveToStash, clicking a Stash slot requests
// MoveToInventory. The Belt is a third, single row (its capacity is fixed
// for the whole game too: InitBelt is only ever called once, at character
// creation, from the one belt item that exists) -- clicking a Belt slot
// requests MoveFromBelt (Belt -> Inventory). The reverse direction
// (Inventory -> Belt) isn't covered: an Inventory slot's click is already
// claimed by MoveToStash, and a second click meaning could only come from
// a distinct gesture (right-click, a mode toggle...) this pass doesn't
// add -- a real, separate follow-up, not an oversight.
//
// ponytail: no real art of any kind exists for this (ROADMAP.md Phase 6),
// so like DevilHUD this renders as plain text/rectangles rather than
// inventing a background frame/icon sheet -- functional over polished.
const (
	devilInventoryColumns = 8
	devilInventoryRows    = 5

	devilInventoryOriginX = 40
	devilInventoryOriginY = 100
	devilStashOriginY     = 250
	devilBeltOriginY      = 400

	devilInventoryColSpacing = 90
	devilInventoryRowSpacing = 20

	devilInventoryCloseButtonX = 40
	devilInventoryCloseButtonY = 70
)

func newDevilInventoryPanel(ui *d2ui.UIManager, l d2util.LogLevel) *devilInventoryPanel {
	p := &devilInventoryPanel{uiManager: ui}

	p.Logger = d2util.NewLogger()
	p.Logger.SetLevel(l)
	p.Logger.SetPrefix(logPrefix)

	return p
}

type devilInventoryPanel struct {
	uiManager           *d2ui.UIManager
	panelGroup          *d2ui.WidgetGroup
	closeButton         *d2ui.Button
	isOpen              bool
	onCloseCb           func()
	onMoveToStashCb     func(inventoryIndex int)
	onMoveToInventoryCb func(stashIndex int)
	onMoveFromBeltCb    func(beltIndex int)

	*d2util.Logger
}

// SetOnCloseCb sets the callback run when the panel is closed.
func (p *devilInventoryPanel) SetOnCloseCb(cb func()) { p.onCloseCb = cb }

// SetOnMoveToStashCb sets the callback run when the player clicks an
// Inventory slot, requesting to move its item to the Stash.
func (p *devilInventoryPanel) SetOnMoveToStashCb(cb func(inventoryIndex int)) { p.onMoveToStashCb = cb }

// SetOnMoveToInventoryCb sets the callback run when the player clicks a
// Stash slot, requesting to move its item to the Inventory.
func (p *devilInventoryPanel) SetOnMoveToInventoryCb(cb func(stashIndex int)) {
	p.onMoveToInventoryCb = cb
}

// SetOnMoveFromBeltCb sets the callback run when the player clicks a Belt
// slot, requesting to move its potion to the Inventory.
func (p *devilInventoryPanel) SetOnMoveFromBeltCb(cb func(beltIndex int)) {
	p.onMoveFromBeltCb = cb
}

// load builds the panel's widgets against player's own Inventory/Stash
// slices -- one inventorySlot per fixed slot index in each (see
// inventorySlot's own doc comment for why this stays correct as content
// changes without ever needing to be rebuilt).
func (p *devilInventoryPanel) load(player *d2mapentity.Player) {
	p.panelGroup = p.uiManager.NewWidgetGroup(d2ui.RenderPriorityInventory)

	p.closeButton = p.uiManager.NewButton(d2ui.ButtonTypeSquareClose, "")
	p.closeButton.SetPosition(devilInventoryCloseButtonX, devilInventoryCloseButtonY)
	p.closeButton.OnActivated(func() { p.Close() })
	p.panelGroup.AddWidget(p.closeButton)

	p.loadGrid(&player.Inventory, devilInventoryOriginY, func(index int) {
		if p.onMoveToStashCb != nil {
			p.onMoveToStashCb(index)
		}
	})

	p.loadGrid(&player.Stash, devilStashOriginY, func(index int) {
		if p.onMoveToInventoryCb != nil {
			p.onMoveToInventoryCb(index)
		}
	})

	p.loadGrid(&player.Belt, devilBeltOriginY, func(index int) {
		if p.onMoveFromBeltCb != nil {
			p.onMoveFromBeltCb(index)
		}
	})

	p.panelGroup.SetVisible(false)
}

// loadGrid builds one devilInventoryColumns x devilInventoryRows grid of
// inventorySlot widgets over container, starting at originY, each one
// requesting onClick(index) when clicked.
func (p *devilInventoryPanel) loadGrid(container *[]string, originY int, onClick func(index int)) {
	for i := 0; i < len(*container); i++ {
		col := i % devilInventoryColumns
		row := i / devilInventoryColumns

		slot := newInventorySlot(p.uiManager, container, i)
		slot.SetPosition(devilInventoryOriginX+col*devilInventoryColSpacing, originY+row*devilInventoryRowSpacing)
		slot.OnActivated(func() { onClick(i) })

		p.panelGroup.AddWidget(slot)
	}
}

// Toggle negates the panel's open state.
func (p *devilInventoryPanel) Toggle() {
	if p.isOpen {
		p.Close()
	} else {
		p.Open()
	}
}

// Open shows the panel.
func (p *devilInventoryPanel) Open() {
	p.isOpen = true
	p.panelGroup.SetVisible(true)
}

// Close hides the panel and runs the close callback, if any.
func (p *devilInventoryPanel) Close() {
	p.isOpen = false
	p.panelGroup.SetVisible(false)

	if p.onCloseCb != nil {
		p.onCloseCb()
	}
}

// IsOpen returns whether the panel is currently shown.
func (p *devilInventoryPanel) IsOpen() bool {
	return p.isOpen
}
