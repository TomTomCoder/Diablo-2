package d2hero

import "errors"

// inventoryCapacity/stashCapacity are placeholder slot counts. Devil's own
// design says storage is "identique à Diablo 2 : stockage volontairement
// limité pour créer des choix de gestion" (devil_game_design_reference.md
// §8: "Inventaire : grille portée sur le personnage (taille contrainte)" /
// "Coffre (stash) : accessible en ville uniquement") -- but gives no exact
// dimensions for Devil's own item set, and D2's real inventory is a
// *spatial* WxH grid where each item's own width/height affects placement.
// Devil's own DevilItemDef has no width/height data (nothing in the design
// specifies it for a magic-only item set of orbs/robes/amulets/rings/
// grimoires), so this models a flat slot count instead of a true spatial
// grid -- same simplification already applied to Belt (devil_belt.go).
// 40 is D2's own real inventory cell count (10x4) cited for a plausible
// scale, not claimed as Devil-specific canon.
const (
	inventoryCapacity = 40
	stashCapacity     = 40
)

// InitStorage sets h's Inventory and Stash to their placeholder capacities,
// discarding whatever was there before. Meant to be called once, at
// character creation -- unlike Belt, storage capacity isn't granted by an
// equipped item, so there's no capacity parameter to pass in.
func (h *HeroState) InitStorage() {
	h.Inventory = make([]string, inventoryCapacity)
	h.Stash = make([]string, stashCapacity)
}

// AddToInventory places itemCode in the first empty inventory slot,
// reporting its index. Errors if every slot is occupied (including if
// InitStorage was never called, leaving a nil/zero-length Inventory).
func (h *HeroState) AddToInventory(itemCode string) (slot int, err error) {
	return addToSlots(h.Inventory, itemCode)
}

// RemoveFromInventory clears inventory slot index and returns what was
// there. Errors for an out-of-range or already-empty slot.
func (h *HeroState) RemoveFromInventory(index int) (itemCode string, err error) {
	return removeFromSlots(h.Inventory, index)
}

// AddToStash places itemCode in the first empty stash slot, reporting its
// index. Errors if every slot is occupied.
//
// ponytail: the design's "accessible en ville uniquement" gate isn't
// enforced here -- there's no "is the player currently in town" state
// tracked anywhere in this engine yet, so that restriction would need to
// live in whatever future packet handler calls this, not in the data model
// itself.
func (h *HeroState) AddToStash(itemCode string) (slot int, err error) {
	return addToSlots(h.Stash, itemCode)
}

// RemoveFromStash clears stash slot index and returns what was there.
// Errors for an out-of-range or already-empty slot.
func (h *HeroState) RemoveFromStash(index int) (itemCode string, err error) {
	return removeFromSlots(h.Stash, index)
}

// addToSlots/removeFromSlots are the shared primitives behind Inventory and
// Stash (and mirror AddPotionToBelt/UsePotion's own belt-slot logic in
// devil_belt.go) -- three independent flat-slot stores with identical
// "first empty slot" placement, not worth three copies of the same loop.
func addToSlots(slots []string, itemCode string) (slot int, err error) {
	for i, occupant := range slots {
		if occupant == "" {
			slots[i] = itemCode
			return i, nil
		}
	}

	return 0, errors.New("no empty slot available")
}

func removeFromSlots(slots []string, index int) (itemCode string, err error) {
	if index < 0 || index >= len(slots) {
		return "", errors.New("slot index out of range")
	}

	itemCode = slots[index]
	if itemCode == "" {
		return "", errors.New("slot is empty")
	}

	slots[index] = ""

	return itemCode, nil
}
