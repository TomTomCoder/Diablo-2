package d2hero

import "errors"

// UseRespecEssence consumes one ItemEssenceDeBossOrdinaire from h's
// inventory and, if found, performs a full respec (RespecComplet) --
// implements a real (if simplified) rare-item gate for Respec complet
// rather than making it freely available, matching the design's own
// intent that it be gated behind something rare (§10) even though the
// exact real trigger it specifies -- "combiner 4 essences de boss dans le
// Cube de Nexus" -- needs boss content that doesn't exist yet. See
// ItemEssenceDeBossOrdinaire's own doc comment for the full reasoning.
//
// Errors (and consumes nothing) if h has no essence in inventory.
func (h *HeroState) UseRespecEssence() (pointsRefunded int, err error) {
	if _, found := h.consumeFromInventory(ItemEssenceDeBossOrdinaire, 1); !found {
		return 0, errors.New("no essence de boss in inventory")
	}

	return h.RespecComplet(), nil
}
