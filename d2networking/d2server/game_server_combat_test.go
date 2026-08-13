package d2server

import (
	"testing"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
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
		connections:         make(map[string]ClientConnection),
		lastCastAt:          make(map[string]time.Time),
		lastMonsterAttackAt: make(map[string]time.Time),
		clock:               time.Now,
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

	if server.canCastNow("p", d2hero.SkillTraitDeFeu) {
		t.Error("expected the cast to be blocked by insufficient mana")
	}
}

func TestCanCastNowDeductsMana(t *testing.T) {
	stats := &d2hero.HeroStatsState{Mana: 10, MaxMana: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	if !server.canCastNow("p", d2hero.SkillTraitDeFeu) {
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
	if !server.canCastNow("p", d2hero.SkillTraitDeFeu) {
		t.Fatal("expected the first cast to succeed")
	}

	if stats.Mana != 0 {
		t.Fatalf("expected 0 mana after spending exactly what was available, got %d", stats.Mana)
	}

	// Not enough time has passed to regen 3 mana at 0 Energy (1/s) -- and
	// we're also still on cooldown, so this should fail regardless.
	server.clock = func() time.Time { return now.Add(baseCastCooldown) }

	if server.canCastNow("p", d2hero.SkillTraitDeFeu) {
		t.Error("expected the second cast to still be blocked (not enough mana regenerated yet)")
	}

	// After 3 more seconds at 1 mana/s (0 Energy), there's enough again.
	server.clock = func() time.Time { return now.Add(baseCastCooldown + 3*time.Second) }

	if !server.canCastNow("p", d2hero.SkillTraitDeFeu) {
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

func TestNearestPlayerPicksClosest(t *testing.T) {
	server := serverWithConnection(nil)
	server.connections["far"] = &fakeClientConnection{
		state: &d2hero.HeroState{X: 100, Y: 100, Stats: &d2hero.HeroStatsState{}},
	}
	server.connections["near"] = &fakeClientConnection{
		state: &d2hero.HeroState{X: 1, Y: 1, Stats: &d2hero.HeroStatsState{}},
	}

	id, pos, found := server.nearestPlayer(d2vector.NewPosition(0, 0))
	if !found {
		t.Fatal("expected to find a nearest player")
	}

	if id != "near" {
		t.Errorf("expected the closer connection 'near', got %q", id)
	}

	if pos.X() != 1 || pos.Y() != 1 {
		t.Errorf("expected position (1,1), got (%v,%v)", pos.X(), pos.Y())
	}
}

func TestNearestPlayerNoConnections(t *testing.T) {
	server := serverWithConnection(nil)

	if _, _, found := server.nearestPlayer(d2vector.NewPosition(0, 0)); found {
		t.Error("expected no nearest player with zero connections")
	}
}

func TestAwardGoldCreditsThePlayer(t *testing.T) {
	state := &d2hero.HeroState{Gold: 0}
	server := serverWithConnection(state)

	server.awardGold("p")

	if state.Gold <= 0 {
		t.Errorf("expected awardGold to credit some gold, got %d", state.Gold)
	}
}

func TestAwardGoldUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the player isn't a connected/resolved state.
	server.awardGold("nobody")
}

func TestAwardExperienceCreditsThePlayer(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Experience: 0, NextLevelExp: 1000}}
	server := serverWithConnection(state)

	server.awardExperience("p")

	if state.Stats.Experience <= 0 {
		t.Errorf("expected awardExperience to credit some experience, got %d", state.Stats.Experience)
	}
}

func TestAwardExperienceUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the player isn't a connected/resolved state.
	server.awardExperience("nobody")
}

func TestAwardExperienceNoStatsNoop(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{Stats: nil})

	// must not panic when the connected player has no Stats yet.
	server.awardExperience("p")
}

func TestResolveNovaHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveNovaHit("nobody", d2hero.SkillNovaDeGivre)
}

func TestKillableNPCsWithinNoMapEnginesReturnsNil(t *testing.T) {
	server := serverWithConnection(nil)

	if got := server.killableNPCsWithin(d2vector.NewPosition(0, 0), novaDeGivreRadiusSubtiles); got != nil {
		t.Errorf("expected nil with no map engines, got %v", got)
	}
}

func TestResolveAoeHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// Boule de feu's path: AoE centered on a targeted position rather than
	// the caster. Must not panic with no map engines to scan.
	server.resolveAoeHit(d2vector.NewPosition(0, 0), "p", d2hero.SkillBouleDeFeu, bouleDeFeuRadiusSubtiles)
}

func TestResolveChampStatiqueHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveChampStatiqueHit("p", d2hero.SkillChampStatique)
}

func TestResolveTelekinesieHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveTelekinesieHit("nobody", d2vector.NewPosition(0, 0))
}

func TestResolveRalentissementHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveRalentissementHit(d2vector.NewPosition(0, 0))
}

func TestAoeAtTargetRadiusSubtilesHasEveryAoeAtTargetSkill(t *testing.T) {
	for _, skillID := range []int{d2hero.SkillBouleDeFeu, d2hero.SkillTempeteStatique} {
		if _, ok := aoeAtTargetRadiusSubtiles[skillID]; !ok {
			t.Errorf("expected skill %d to be dispatched via aoeAtTargetRadiusSubtiles", skillID)
		}
	}
}

func TestTryMonsterAttackAppliesDamageAndGatesOnCooldown(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 10-monsterAttackDamage {
		t.Errorf("expected HP %d after one attack, got %d", 10-monsterAttackDamage, stats.Health)
	}

	// same monster, still on cooldown: no second hit
	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 10-monsterAttackDamage {
		t.Errorf("expected no further damage while on cooldown, got HP %d", stats.Health)
	}

	// a different monster has its own independent cooldown
	server.tryMonsterAttack("npc-2", "p")

	if stats.Health != 10-2*monsterAttackDamage {
		t.Errorf("expected a second monster's attack to land, got HP %d", stats.Health)
	}

	// after the cooldown elapses, npc-1 can attack again
	server.clock = func() time.Time { return now.Add(monsterAttackCooldown) }
	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 10-3*monsterAttackDamage {
		t.Errorf("expected npc-1's attack to land after its cooldown elapsed, got HP %d", stats.Health)
	}
}

func TestResolveAttackDamageTraitDeFeu(t *testing.T) {
	// Trait de feu has its own base_sort (6), distinct from the flat
	// fallback (4) -- proves per-skill data actually takes effect.
	const traitDeFeuBase = 6

	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 0}})

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != traitDeFeuBase {
		t.Errorf("expected Trait de feu base damage %d, got %d", traitDeFeuBase, got)
	}

	// and it still scales with Energy on top of its own base
	server = serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 100}})

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != traitDeFeuBase*2 {
		t.Errorf("expected Trait de feu at 100 Energy to be %d, got %d", traitDeFeuBase*2, got)
	}
}

func TestResolveAttackDamageWeaponFireModifier(t *testing.T) {
	const traitDeFeuBase = 6 // no Energy scaling in this test (Energy: 0)

	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: 0},
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	})

	want := traitDeFeuBase + (traitDeFeuBase*10)/100 // Bâton de l'Apprenti: +10% dégâts Feu

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != want {
		t.Errorf("expected the equipped staff's Fire modifier applied (%d), got %d", want, got)
	}
}

func TestResolveAttackDamageNoWeaponNoModifier(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 0}})

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != 6 {
		t.Errorf("expected no modifier with nothing equipped, got %d", got)
	}
}
