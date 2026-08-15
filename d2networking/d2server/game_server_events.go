package d2server

import "time"

// Devil's temporary drop-boost events (devil_game_design_reference.md §9):
// "Pour compenser la rareté des objets de hauts TC sans supprimer le
// sentiment de récompense, Devil intègre un système d'Events."
//
// Only the two duration-based events (a temporary buff that's simply
// active or not) fit ActivateEvent/IsEventActive's shape. "Boss corrompu"
// (a rare monster spawn) and "Quête d'urgence" (triggered by progression)
// aren't timed buffs -- they're one-shot effects tied to a spawn/quest
// system that doesn't exist yet, so they're not modeled here.
const (
	// EventTempeteDeLoot: "NoDrop divisé par 2 sur toute la session".
	EventTempeteDeLoot = "tempete_de_loot"
	// EventNuitDeLApocalypse: "Multiplicateur x1.5 sur le Magic Find".
	EventNuitDeLApocalypse = "nuit_de_lapocalypse"
)

// tempeteDeLootNoDropDivisor/nuitDeLApocalypseMagicFindPercent are the
// effect magnitudes the design specifies for each event.
//
// ponytail: declared for when an actual Treasure Class roll consults
// these -- Devil has no drop engine of its own yet to plug them into (see
// ROADMAP.md Phase 5's Treasure Classes/inventory notes), so nothing reads
// these two constants yet either.
const (
	tempeteDeLootNoDropDivisor        = 2
	nuitDeLApocalypseMagicFindPercent = 150 // x1.5
)

// ActivateEvent marks eventName active for duration from now. Reactivating
// an already-active event simply resets its expiry from now, rather than
// stacking durations.
func (g *GameServer) ActivateEvent(eventName string, duration time.Duration) {
	g.Lock()
	defer g.Unlock()

	g.activeEventUntil[eventName] = g.clock().Add(duration)
}

// IsEventActive reports whether eventName is currently active.
func (g *GameServer) IsEventActive(eventName string) bool {
	g.Lock()
	defer g.Unlock()

	until, ok := g.activeEventUntil[eventName]

	return ok && g.clock().Before(until)
}
