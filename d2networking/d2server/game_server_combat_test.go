package d2server

import (
	"testing"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2clientconnectiontype"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket"
)

// fakeClientConnection is a minimal ClientConnection test double: only
// GetPlayerState is exercised by resolveAttackDamage.
type fakeClientConnection struct {
	state *d2hero.HeroState
}

func (f *fakeClientConnection) GetUniqueID() string { return "" }
func (f *fakeClientConnection) GetConnectionType() d2clientconnectiontype.ClientConnectionType {
	return d2clientconnectiontype.LANClient
}
func (f *fakeClientConnection) SendPacketToClient(_ d2netpacket.NetPacket) error { return nil }
func (f *fakeClientConnection) GetPlayerState() *d2hero.HeroState                { return f.state }
func (f *fakeClientConnection) SetPlayerState(state *d2hero.HeroState)           { f.state = state }

func serverWithConnection(state *d2hero.HeroState) *GameServer {
	server := &GameServer{
		connections: make(map[string]ClientConnection),
		lastCastAt:  make(map[string]time.Time),
		clock:       time.Now,
	}
	if state != nil {
		server.connections["p"] = &fakeClientConnection{state: state}
	}

	return server
}

// unknownSkillID is any ID not present in skillBaseSortDamage, so tests
// that aren't about per-skill data get the flat baseSortDamage fallback.
const unknownSkillID = -1

func TestResolveAttackDamageFallbacks(t *testing.T) {
	cases := map[string]*d2hero.HeroState{
		"no connection at all": nil,
		"nil stats":            {Stats: nil},
	}

	for name, state := range cases {
		server := serverWithConnection(state)

		if got := server.resolveAttackDamage("p", unknownSkillID); got != baseSortDamage {
			t.Errorf("%s: expected fallback %d, got %d", name, baseSortDamage, got)
		}
	}
}

func TestResolveAttackDamageEnergyScaling(t *testing.T) {
	// Dégâts = base_sort × (1 + Energy / 100), integer division.
	cases := []struct {
		energy   int
		expected int
	}{
		{energy: 0, expected: baseSortDamage},                              // no scaling
		{energy: 100, expected: baseSortDamage * 2},                        // double damage
		{energy: 50, expected: baseSortDamage + (baseSortDamage*50)/100},   // +50%
		{energy: 200, expected: baseSortDamage + (baseSortDamage*200)/100}, // +200%
	}

	for _, c := range cases {
		server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: c.energy}})

		if got := server.resolveAttackDamage("p", unknownSkillID); got != c.expected {
			t.Errorf("energy=%d: expected %d, got %d", c.energy, c.expected, got)
		}
	}
}

func TestCastCooldownForScalesWithDexterity(t *testing.T) {
	cases := []struct {
		dexterity int
		want      time.Duration
	}{
		{dexterity: 0, want: baseCastCooldown},
		{dexterity: 25, want: baseCastCooldown / 2},
		{dexterity: 1000, want: minCastCooldown}, // floors, never reaches zero
	}

	for _, c := range cases {
		if got := castCooldownFor(c.dexterity); got != c.want {
			t.Errorf("dexterity=%d: expected %v, got %v", c.dexterity, c.want, got)
		}
	}
}

// plentyOfMana is enough mana for these cooldown-focused tests to never be
// blocked by manaCostFor -- that's covered separately in the mana tests.
var plentyOfMana = &d2hero.HeroStatsState{Mana: 1000, MaxMana: 1000}

func TestCanCastNowGatesOnCooldown(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{
		Dexterity: 0, Mana: plentyOfMana.Mana, MaxMana: plentyOfMana.MaxMana,
	}})

	now := time.Now()
	server.clock = func() time.Time { return now }

	if !server.canCastNow("p", unknownSkillID) {
		t.Fatal("first cast should always be allowed")
	}

	if server.canCastNow("p", unknownSkillID) {
		t.Error("second cast at the same instant should be gated by cooldown")
	}

	server.clock = func() time.Time { return now.Add(baseCastCooldown) }

	if !server.canCastNow("p", unknownSkillID) {
		t.Error("cast after the cooldown elapsed should be allowed")
	}
}

func TestCanCastNowIsPerCaster(t *testing.T) {
	server := serverWithConnection(nil)
	server.connections["a"] = &fakeClientConnection{state: &d2hero.HeroState{Stats: plentyOfMana}}
	server.connections["b"] = &fakeClientConnection{state: &d2hero.HeroState{Stats: plentyOfMana}}

	now := time.Now()
	server.clock = func() time.Time { return now }

	if !server.canCastNow("a", unknownSkillID) {
		t.Fatal("a's first cast should be allowed")
	}

	if !server.canCastNow("b", unknownSkillID) {
		t.Error("b should have their own independent cooldown from a")
	}
}

func TestCanCastNowGatesOnMana(t *testing.T) {
	// Trait de feu costs 3 mana (skillManaCost); starting with only 2 isn't
	// enough, and no time has passed for regen to help.
	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Mana: 2, MaxMana: 10}})

	if server.canCastNow("p", skillTraitDeFeu) {
		t.Error("expected the cast to be blocked by insufficient mana")
	}
}

func TestCanCastNowDeductsMana(t *testing.T) {
	stats := &d2hero.HeroStatsState{Mana: 10, MaxMana: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	if !server.canCastNow("p", skillTraitDeFeu) {
		t.Fatal("expected the cast to succeed with enough mana")
	}

	if stats.Mana != 7 {
		t.Errorf("expected mana to drop by the skill's cost (3), got %d", stats.Mana)
	}
}

func TestCanCastNowRegeneratesManaOverTime(t *testing.T) {
	stats := &d2hero.HeroStatsState{Energy: 0, Mana: 3, MaxMana: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	now := time.Now()
	server.clock = func() time.Time { return now }

	// Spend down to 0 mana with the first cast (cost 3, had exactly 3).
	if !server.canCastNow("p", skillTraitDeFeu) {
		t.Fatal("expected the first cast to succeed")
	}

	if stats.Mana != 0 {
		t.Fatalf("expected 0 mana after spending exactly what was available, got %d", stats.Mana)
	}

	// Not enough time has passed to regen 3 mana at 0 Energy (1/s) -- and
	// we're also still on cooldown, so this should fail regardless.
	server.clock = func() time.Time { return now.Add(baseCastCooldown) }

	if server.canCastNow("p", skillTraitDeFeu) {
		t.Error("expected the second cast to still be blocked (not enough mana regenerated yet)")
	}

	// After 3 more seconds at 1 mana/s (0 Energy), there's enough again.
	server.clock = func() time.Time { return now.Add(baseCastCooldown + 3*time.Second) }

	if !server.canCastNow("p", skillTraitDeFeu) {
		t.Error("expected the cast to succeed once enough mana regenerated")
	}
}

func TestManaRegenPerSecondScalesWithEnergy(t *testing.T) {
	if got := manaRegenPerSecond(0); got != baseManaRegenPerSecond {
		t.Errorf("at 0 Energy, expected base regen %v, got %v", baseManaRegenPerSecond, got)
	}

	if got := manaRegenPerSecond(50); got <= baseManaRegenPerSecond {
		t.Errorf("higher Energy should regen faster than the base rate, got %v", got)
	}
}

func TestResolveAttackDamageTraitDeFeu(t *testing.T) {
	// Trait de feu has its own base_sort (6), distinct from the flat
	// fallback (4) -- proves per-skill data actually takes effect.
	const traitDeFeuBase = 6

	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 0}})

	if got := server.resolveAttackDamage("p", skillTraitDeFeu); got != traitDeFeuBase {
		t.Errorf("expected Trait de feu base damage %d, got %d", traitDeFeuBase, got)
	}

	// and it still scales with Energy on top of its own base
	server = serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 100}})

	if got := server.resolveAttackDamage("p", skillTraitDeFeu); got != traitDeFeuBase*2 {
		t.Errorf("expected Trait de feu at 100 Energy to be %d, got %d", traitDeFeuBase*2, got)
	}
}
