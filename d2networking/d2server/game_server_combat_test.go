package d2server

import (
	"testing"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2clientconnectiontype"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket"
)

// fakeClientConnection is a minimal ClientConnection test double: only
// GetPlayerState is exercised by resolveAttackDamage. sent records every
// packet handed to SendPacketToClient, for tests that need to inspect what
// was broadcast (e.g. TestBroadcastNPCMovedSendsCurrentPosition).
type fakeClientConnection struct {
	state *d2hero.HeroState
	sent  []d2netpacket.NetPacket
}

func (f *fakeClientConnection) GetUniqueID() string { return "" }
func (f *fakeClientConnection) GetConnectionType() d2clientconnectiontype.ClientConnectionType {
	return d2clientconnectiontype.LANClient
}
func (f *fakeClientConnection) SendPacketToClient(p d2netpacket.NetPacket) error {
	f.sent = append(f.sent, p)
	return nil
}
func (f *fakeClientConnection) GetPlayerState() *d2hero.HeroState      { return f.state }
func (f *fakeClientConnection) SetPlayerState(state *d2hero.HeroState) { f.state = state }

func serverWithConnection(state *d2hero.HeroState) *GameServer {
	server := &GameServer{
		connections:         make(map[string]ClientConnection),
		lastCastAt:          make(map[string]time.Time),
		lastMonsterAttackAt: make(map[string]time.Time),
		lastTranscendanceAt: make(map[string]time.Time),
		activeEventUntil:    make(map[string]time.Time),
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
var plentyOfMana = &d2hero.HeroStatsState{Mana: 1000, MaxMana: 1000} // nolint:gochecknoglobals // test fixture, not shared runtime state

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

// TestBroadcastPlayerManaSendsCurrentMana is a regression test:
// canCastNow spends and regenerates Mana entirely server-side (even on a
// cast that ultimately fails for lack of mana), but nothing ever broadcast
// the result -- a client's own mana orb would never move no matter how
// many spells were cast.
func TestBroadcastPlayerManaSendsCurrentMana(t *testing.T) {
	stats := &d2hero.HeroStatsState{Mana: 7, MaxMana: 10}
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: stats}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	server.broadcastPlayerMana("p")

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	used, err := d2netpacket.UnmarshalPotionUsed(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PotionUsedPacket: %v", err)
	}

	if used.Mana != 7 {
		t.Errorf("expected the broadcast to carry the current Mana (7), got %d", used.Mana)
	}
}

func TestBroadcastPlayerManaUnknownPlayerNoop(t *testing.T) {
	server := &GameServer{connections: map[string]ClientConnection{}}

	// must not panic when the caster isn't a connected/resolved player.
	server.broadcastPlayerMana("nobody")
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

// TestCanCastNowRegenerationAccelereeSpeedsUpRegen is a regression test for
// Régénération accélérée ("Augmente la vitesse de régénération du mana",
// devil_game_design_reference.md §7). Uses a passive (ManaCost 0) as the
// probe cast so canCastNow's own mana deduction doesn't muddy the regen
// amount being checked.
func TestCanCastNowRegenerationAccelereeSpeedsUpRegen(t *testing.T) {
	const points = 2

	stats := &d2hero.HeroStatsState{Energy: 0, Mana: 0, MaxMana: 100}
	state := &d2hero.HeroState{
		Stats:  stats,
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillRegenerationAcceleree: {SkillPoints: points}},
	}
	server := serverWithConnection(state)

	now := time.Now()
	server.clock = func() time.Time { return now }
	server.lastCastAt["p"] = now

	const elapsedSeconds = 10.0
	server.clock = func() time.Time { return now.Add(time.Duration(elapsedSeconds) * time.Second) }

	baseRegen := int(manaRegenPerSecond(0) * elapsedSeconds)
	percent := d2hero.RegenerationAccellereePercent(points)
	boostedRegen := int(manaRegenPerSecond(0) * (1 + float64(percent)/100) * elapsedSeconds)

	if boostedRegen <= baseRegen {
		t.Fatal("test setup error: chosen points don't make the bonus visible after integer truncation")
	}

	if !server.canCastNow("p", d2hero.SkillMaitriseElementaire) {
		t.Fatal("expected the (free, ManaCost 0) cast to succeed")
	}

	if stats.Mana != boostedRegen {
		t.Errorf("expected Régénération accélérée's boosted regen of %d, got %d", boostedRegen, stats.Mana)
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

func TestResolveSelfCenteredAoeHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveSelfCenteredAoeHit("nobody", d2hero.SkillNovaDeGivre, novaDeGivreRadiusSubtiles)
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

// TestDropLootNoMapEnginesDoesNotPanic mirrors
// TestResolveChampStatiqueHitNoMapEnginesDoesNotPanic: dropLoot's own
// len(g.mapEngines) == 0 guard returns before ever touching npc, so a nil
// *d2mapentity.NPC is safe here -- constructing a real one needs a full
// MPQ-loaded AssetManager, the same limitation documented throughout this
// package's other tests.
func TestDropLootNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	server.dropLoot(nil)
}

// TestBroadcastNPCMovedSendsCurrentPosition is a regression test: Knockback
// (Télékinésie) and Pull (Vortex) used to mutate an NPC's position with no
// way for clients to find out, so the entity would silently stay put on
// their side. Constructs an NPC directly (Position is a promoted exported
// field, unlike the private monstat-backed fields other tests need a real
// factory for) since broadcastNPCMoved only ever reads ID()/GetPosition().
func TestBroadcastNPCMovedSendsCurrentPosition(t *testing.T) {
	conn := &fakeClientConnection{}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	npc := &d2mapentity.NPC{}
	npc.Position.Set(3, 4)

	server.broadcastNPCMoved(npc)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	moved, err := d2netpacket.UnmarshalNPCMoved(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal NPCMovedPacket: %v", err)
	}

	if moved.X != 3 || moved.Y != 4 {
		t.Errorf("expected position (3, 4), got (%v, %v)", moved.X, moved.Y)
	}
}

// TestBroadcastNPCStatusEffectSendsEffectAndExpiry is a regression test:
// Éclat de glace/Ralentissement/Distorsion temporelle's slow, Prison de
// glace's immobilize, Amplification's amplify, and Rupture arcane's
// resistance strip all used to mutate the server's own copy of the NPC with
// no way for clients to find out.
func TestBroadcastNPCStatusEffectSendsEffectAndExpiry(t *testing.T) {
	conn := &fakeClientConnection{}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	npc := &d2mapentity.NPC{}
	until := time.Now().Add(4 * time.Second)

	server.broadcastNPCStatusEffect(npc, d2netpacket.NPCStatusSlowed, until)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalNPCStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal NPCStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.NPCStatusSlowed {
		t.Errorf("expected effect %q, got %q", d2netpacket.NPCStatusSlowed, status.Effect)
	}

	if !status.Until.Equal(until) {
		t.Errorf("expected expiry %v, got %v", until, status.Until)
	}
}

func TestResolveTelekinesieHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveTelekinesieHit("nobody", d2vector.NewPosition(0, 0))
}

func TestAmplifiedDamage(t *testing.T) {
	if got := amplifiedDamage(10, false); got != 10 {
		t.Errorf("expected unamplified damage unchanged, got %d", got)
	}

	if got, want := amplifiedDamage(10, true), 10*amplificationDamagePercent/100; got != want {
		t.Errorf("expected amplified damage %d, got %d", want, got)
	}
}

func TestOverloadBonusDamage(t *testing.T) {
	if got := overloadBonusDamage(200, false); got != 0 {
		t.Errorf("expected no bonus damage without Overload active, got %d", got)
	}

	if got, want := overloadBonusDamage(200, true), 200*overloadDamagePercent/100; got != want {
		t.Errorf("expected %d%% of current HP (%d) as bonus damage, got %d", overloadDamagePercent, want, got)
	}
}

func TestShouldIgniteBurn(t *testing.T) {
	if shouldIgniteBurn(1000, false) {
		t.Error("expected no ignition without MarqueArdenteActive")
	}

	if !shouldIgniteBurn(1000, true) {
		t.Error("expected a real hit (nonzero skillID) to ignite while active")
	}

	// The regression this pins down: a tick's own recursive call into
	// applyResolvedDamage (skillID == burnTickSkillID) must never
	// re-ignite the burn it's currently ticking down, or the burn would
	// refresh itself forever instead of expiring.
	if shouldIgniteBurn(burnTickSkillID, true) {
		t.Error("expected a burn tick's own damage application not to re-ignite the burn")
	}
}

// advanceBurningNPC tests use a bare, non-killable *d2mapentity.NPC (no
// exported constructor exists for a real killable one outside its own
// package -- same documented limitation as TestDropLootNoMapEnginesDoesNotPanic
// above). That's enough to test the tick-scheduling logic in
// advanceBurningNPC itself: the actual HP reduction inside
// applyResolvedDamage->ApplyDamage is already covered by
// TestNPCApplyDamageNotKillable/TestNPCApplyDamage in d2mapentity's own
// tests.
func TestAdvanceBurningNPCNotBurningIsNoop(t *testing.T) {
	server := serverWithConnection(nil)
	npc := &d2mapentity.NPC{}

	server.advanceBurningNPC(npc)

	if !npc.NextBurnTickAt.IsZero() {
		t.Error("expected a never-burning NPC's NextBurnTickAt to stay zero")
	}
}

func TestAdvanceBurningNPCNotYetDueIsNoop(t *testing.T) {
	server := serverWithConnection(nil)
	now := time.Now()
	server.clock = func() time.Time { return now }

	npc := &d2mapentity.NPC{}
	npc.ApplyBurn("attacker1", now.Add(5*time.Second), 20, now.Add(time.Second))

	server.advanceBurningNPC(npc)

	if !npc.NextBurnTickAt.Equal(now.Add(time.Second)) {
		t.Error("expected NextBurnTickAt to stay untouched before it's actually due")
	}
}

func TestAdvanceBurningNPCAdvancesNextTickWhenDue(t *testing.T) {
	server := serverWithConnection(nil)
	now := time.Now()

	npc := &d2mapentity.NPC{}
	npc.ApplyBurn("attacker1", now.Add(5*time.Second), 20, now)

	server.clock = func() time.Time { return now }
	server.advanceBurningNPC(npc)

	if !npc.NextBurnTickAt.Equal(now.Add(marqueArdenteTickInterval)) {
		t.Errorf("expected NextBurnTickAt advanced to %v, got %v", now.Add(marqueArdenteTickInterval), npc.NextBurnTickAt)
	}
}

func TestResolveTeleportationHitMovesThePlayer(t *testing.T) {
	state := &d2hero.HeroState{X: 0, Y: 0}
	server := serverWithConnection(state)

	server.resolveTeleportationHit("p", d2vector.NewPosition(10, 20))

	if state.X != 10 || state.Y != 20 {
		t.Errorf("expected player moved to (10, 20), got (%v, %v)", state.X, state.Y)
	}
}

func TestResolveTeleportationHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveTeleportationHit("nobody", d2vector.NewPosition(10, 20))
}

func TestResolveRalentissementHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveRalentissementHit(d2vector.NewPosition(0, 0))
}

func TestResolvePrisonDeGlaceHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolvePrisonDeGlaceHit(d2vector.NewPosition(0, 0))
}

func TestResolveVortexHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveVortexHit(d2vector.NewPosition(0, 0))
}

func TestResolveDistorsionTemporelleHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveDistorsionTemporelleHit()
}

func TestResolveApocalypseHitNoMapEnginesDoesNotPanic(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when there's no map engine to scan.
	server.resolveApocalypseHit("p", d2hero.SkillApocalypse)
}

func TestResolveBouclierDeManaHitTogglesManaShield(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}
	server := serverWithConnection(state)

	if state.Stats.ManaShieldActive {
		t.Fatal("expected ManaShieldActive to start false")
	}

	server.resolveBouclierDeManaHit("p")

	if !state.Stats.ManaShieldActive {
		t.Error("expected first cast to turn ManaShieldActive on")
	}

	server.resolveBouclierDeManaHit("p")

	if state.Stats.ManaShieldActive {
		t.Error("expected second cast to turn ManaShieldActive back off")
	}
}

func TestResolveBouclierDeManaHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveBouclierDeManaHit("nobody")
}

// TestResolveBouclierDeManaHitBroadcastsToggle is a regression test:
// resolveBouclierDeManaHit used to toggle ManaShieldActive server-side with
// nothing telling clients (even the caster's own, since solo play still
// round-trips through a local client/server -- ROADMAP.md).
func TestResolveBouclierDeManaHitBroadcastsToggle(t *testing.T) {
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	server.resolveBouclierDeManaHit("p")

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalPlayerStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.PlayerStatusManaShield || !status.Active {
		t.Errorf("expected effect %q active=true, got %q active=%v", d2netpacket.PlayerStatusManaShield, status.Effect, status.Active)
	}
}

func TestAoeAtTargetRadiusSubtilesHasEveryAoeAtTargetSkill(t *testing.T) {
	for _, skillID := range []int{d2hero.SkillBouleDeFeu, d2hero.SkillTempeteStatique, d2hero.SkillOrbeGlaciale, d2hero.SkillMeteore} {
		if _, ok := aoeAtTargetRadiusSubtiles[skillID]; !ok {
			t.Errorf("expected skill %d to be dispatched via aoeAtTargetRadiusSubtiles", skillID)
		}
	}
}

func TestSelfCenteredAoeRadiusSubtilesHasEverySelfCenteredAoeSkill(t *testing.T) {
	for _, skillID := range []int{d2hero.SkillNovaDeGivre, d2hero.SkillTempeteDeLames} {
		if _, ok := selfCenteredAoeRadiusSubtiles[skillID]; !ok {
			t.Errorf("expected skill %d to be dispatched via selfCenteredAoeRadiusSubtiles", skillID)
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

// TestTryMonsterAttackWithManaShieldBroadcastsDrainedMana is a regression
// test: Bouclier de mana drains Mana instead of Health while active
// (HeroStatsState.ApplyDamageWithManaShield), but the broadcast packet
// only ever carried HP -- the client's own Mana never learned it was
// drained.
func TestTryMonsterAttackWithManaShieldBroadcastsDrainedMana(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10, Mana: 20, MaxMana: 20, ManaShieldActive: true}
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: stats}}
	server := &GameServer{
		connections:         map[string]ClientConnection{"p": conn},
		lastMonsterAttackAt: make(map[string]time.Time),
		clock:               time.Now,
	}

	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 10 {
		t.Fatalf("expected Health untouched while Bouclier de mana absorbs the hit, got %d", stats.Health)
	}

	if stats.Mana != 20-monsterAttackDamage {
		t.Fatalf("expected Mana drained by the hit, got %d", stats.Mana)
	}

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	damaged, err := d2netpacket.UnmarshalPlayerDamaged(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerDamagedPacket: %v", err)
	}

	if damaged.Mana != stats.Mana {
		t.Errorf("expected the broadcast to carry the drained Mana (%d), got %d", stats.Mana, damaged.Mana)
	}
}

func TestTryMonsterAttackWithArmureDeGlaceReducesDamage(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10, ArmureDeGlaceActive: true}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.tryMonsterAttack("npc-1", "p")

	reduced := monsterAttackDamage * (100 - armureDeGlaceDamageReductionPercent) / 100
	if got, want := 10-stats.Health, reduced; got != want {
		t.Errorf("expected Armure de glace to reduce damage to %d, got %d", want, got)
	}

	if reduced >= monsterAttackDamage {
		t.Fatal("test is meaningless if the reduction doesn't actually reduce anything")
	}
}

// TestTryMonsterAttackWithBouclierDeManaSynergyBoostsArmureDeGlace is a
// regression test for Bouclier de mana's own synergy ("chaque point dans
// Bouclier de mana augmente l'absorption de Armure de glace",
// devil_game_design_reference.md §7) -- applies even though the shield
// itself isn't toggled on, since the design only requires the points to be
// invested.
func TestTryMonsterAttackWithBouclierDeManaSynergyBoostsArmureDeGlace(t *testing.T) {
	const points = 3

	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10, ArmureDeGlaceActive: true}
	state := &d2hero.HeroState{
		Stats:  stats,
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillBouclierDeMana: {SkillPoints: points}},
	}
	server := serverWithConnection(state)

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.tryMonsterAttack("npc-1", "p")

	percent := armureDeGlaceDamageReductionPercent + d2hero.BouclierDeManaSynergyReductionPercent(points)
	reduced := monsterAttackDamage * (100 - percent) / 100

	if got, want := 10-stats.Health, reduced; got != want {
		t.Errorf("expected the synergy to boost Armure de glace's reduction to %d, got %d", want, got)
	}

	if reduced >= monsterAttackDamage*(100-armureDeGlaceDamageReductionPercent)/100 {
		t.Fatal("test is meaningless if the synergy doesn't actually reduce damage further")
	}
}

// TestTryMonsterAttackWithBouclierDeManaSynergyCapsAtZeroDamage is a
// regression test for a correctness bug: with enough points invested in
// Bouclier de mana (no per-skill point cap exists anywhere), the combined
// reduction percent could exceed 100, making (100-percent) negative --
// damage would go negative and heal the player instead of hurting them.
func TestTryMonsterAttackWithBouclierDeManaSynergyCapsAtZeroDamage(t *testing.T) {
	const points = 100 // armureDeGlaceDamageReductionPercent(30) + 100*2 is far past 100%

	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10, ArmureDeGlaceActive: true}
	state := &d2hero.HeroState{
		Stats:  stats,
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillBouclierDeMana: {SkillPoints: points}},
	}
	server := serverWithConnection(state)

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.tryMonsterAttack("npc-1", "p")

	if stats.Health > 10 {
		t.Fatalf("expected damage to floor at 0 rather than heal the player, got Health %d (started at 10)", stats.Health)
	}

	if stats.Health != 10 {
		t.Errorf("expected zero damage taken once the reduction is capped at 100%%, got Health %d", stats.Health)
	}
}

func TestTryMonsterAttackWithTranscendanceSavesFromDeath(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 1, MaxHealth: 10}
	state := &d2hero.HeroState{
		Stats:  stats,
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTranscendance: {}},
	}
	server := serverWithConnection(state)

	now := time.Now()
	server.clock = func() time.Time { return now }

	// a hit that would otherwise kill (1 HP, more damage than that incoming)
	server.tryMonsterAttack("npc-1", "p")

	if stats.Health <= 0 {
		t.Fatalf("expected Transcendance to save the Mage from death, got HP %d", stats.Health)
	}

	if stats.Health != stats.MaxHealth*transcendanceHealPercent/100 {
		t.Errorf("expected the death-save heal to restore %d%% of MaxHealth, got HP %d",
			transcendanceHealPercent, stats.Health)
	}
}

func TestTryMonsterAttackWithoutTranscendanceStillDies(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 1, MaxHealth: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 0 {
		t.Errorf("expected no death-save without the skill learned, got HP %d", stats.Health)
	}
}

func TestTryTranscendGatesOnCooldown(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 1, MaxHealth: 10}
	state := &d2hero.HeroState{
		Stats:  stats,
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTranscendance: {}},
	}
	server := serverWithConnection(state)

	now := time.Now()
	server.clock = func() time.Time { return now }

	if !server.tryTranscend("p", state) {
		t.Fatal("expected the first transcend to succeed")
	}

	stats.Health = 0 // simulate dying again immediately

	if server.tryTranscend("p", state) {
		t.Error("expected a second transcend to be gated by its own cooldown")
	}

	if stats.Health != 0 {
		t.Error("expected no heal from a transcend attempt blocked by cooldown")
	}

	server.clock = func() time.Time { return now.Add(transcendanceCooldown) }

	if !server.tryTranscend("p", state) {
		t.Error("expected transcend to succeed again once its cooldown elapsed")
	}
}

func TestResolveBouclierDeManaHitAndArmureDeGlaceAreIndependentToggles(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}
	server := serverWithConnection(state)

	server.resolveArmureDeGlaceHit("p")

	if !state.Stats.ArmureDeGlaceActive {
		t.Error("expected first cast to turn ArmureDeGlaceActive on")
	}

	if state.Stats.ManaShieldActive {
		t.Error("expected Armure de glace to not affect ManaShieldActive")
	}

	server.resolveArmureDeGlaceHit("p")

	if state.Stats.ArmureDeGlaceActive {
		t.Error("expected second cast to turn ArmureDeGlaceActive back off")
	}
}

func TestResolveArmureDeGlaceHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveArmureDeGlaceHit("nobody")
}

// TestResolveArmureDeGlaceHitBroadcastsToggle is the same regression as
// resolveBouclierDeManaHit's, for the other toggle.
func TestResolveArmureDeGlaceHitBroadcastsToggle(t *testing.T) {
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	server.resolveArmureDeGlaceHit("p")

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalPlayerStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.PlayerStatusArmureDeGlace || !status.Active {
		t.Errorf("expected effect %q active=true, got %q active=%v", d2netpacket.PlayerStatusArmureDeGlace, status.Effect, status.Active)
	}
}

func TestNpcByIDNoMapEnginesReturnsNil(t *testing.T) {
	server := serverWithConnection(nil)

	if got := server.npcByID("npc-1"); got != nil {
		t.Errorf("expected nil with no map engines, got %v", got)
	}
}

func TestResolveEveilDuNexusHitGrantsImmunityAndHeals(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 2, MaxHealth: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.resolveEveilDuNexusHit("p")

	if !stats.IsMagicImmune(now) {
		t.Error("expected Éveil du Nexus to grant immediate magic immunity")
	}

	if want := 2 + 10*eveilDuNexusHealPercent/100; stats.Health != want {
		t.Errorf("expected Health %d after the heal, got %d", want, stats.Health)
	}
}

func TestTryMonsterAttackSkipsDamageWhenMagicImmune(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 10, MaxHealth: 10}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	now := time.Now()
	server.clock = func() time.Time { return now }

	server.resolveEveilDuNexusHit("p")
	server.tryMonsterAttack("npc-1", "p")

	if stats.Health != 10 {
		t.Errorf("expected magic immunity to block all damage, got HP %d", stats.Health)
	}
}

func TestResolveEveilDuNexusHitUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveEveilDuNexusHit("nobody")
}

// TestResolveEveilDuNexusHitBroadcastsImmunityAndHeal is a regression test:
// the immunity and the heal it grants were both previously invisible to
// clients -- the heal used PlayerDamagedPacket (it just carries current HP,
// doubling fine for a heal) rather than a new packet type.
func TestResolveEveilDuNexusHitBroadcastsImmunityAndHeal(t *testing.T) {
	stats := &d2hero.HeroStatsState{Health: 2, MaxHealth: 10}
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: stats}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}, clock: time.Now}

	server.resolveEveilDuNexusHit("p")

	if len(conn.sent) != 2 {
		t.Fatalf("expected 2 packets sent (immunity + heal), got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalPlayerStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.PlayerStatusMagicImmune || !status.Active {
		t.Errorf("expected effect %q active=true, got %q active=%v", d2netpacket.PlayerStatusMagicImmune, status.Effect, status.Active)
	}

	damaged, err := d2netpacket.UnmarshalPlayerDamaged(conn.sent[1].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerDamagedPacket: %v", err)
	}

	if damaged.HP != stats.Health {
		t.Errorf("expected the heal broadcast to carry the post-heal HP (%d), got %d", stats.Health, damaged.HP)
	}
}

func TestResolveUsePotionRestoresManaAndClearsSlot(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Mana: 0, MaxMana: 100}}
	state.InitBelt(2)

	if _, err := state.AddPotionToBelt(d2hero.ItemPotionDeMana); err != nil {
		t.Fatalf("test setup: AddPotionToBelt failed: %v", err)
	}

	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateUsePotionRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateUsePotionRequestPacket failed: %v", err)
	}

	server.resolveUsePotion(packet)

	if want := d2hero.ItemManaRestoreAmount(d2hero.ItemPotionDeMana); state.Stats.Mana != want {
		t.Errorf("expected Mana restored to %d, got %d", want, state.Stats.Mana)
	}

	if state.Belt[0] != "" {
		t.Errorf("expected belt slot 0 cleared after use, got %q", state.Belt[0])
	}
}

func TestResolveUsePotionOnEmptySlotIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Mana: 0, MaxMana: 100}}
	state.InitBelt(2)

	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateUsePotionRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateUsePotionRequestPacket failed: %v", err)
	}

	server.resolveUsePotion(packet)

	if state.Stats.Mana != 0 {
		t.Errorf("expected no Mana restored from an empty slot, got %d", state.Stats.Mana)
	}
}

func TestResolveUsePotionUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateUsePotionRequestPacket("nobody", 0)
	if err != nil {
		t.Fatalf("test setup: CreateUsePotionRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveUsePotion(packet)
}

func TestResolveCraftUpgradesWeaponAndBroadcastsResult(t *testing.T) {
	state := &d2hero.HeroState{
		Gold: 100,
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	}
	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateCraftRequestPacket("p", d2hero.RecipeUpgradeBatonApprenti)
	if err != nil {
		t.Fatalf("test setup: CreateCraftRequestPacket failed: %v", err)
	}

	server.resolveCraft(packet)

	if state.Equipment.RightHand.ItemCode != d2hero.ItemBatonInitie {
		t.Errorf("expected weapon upgraded to %q, got %q", d2hero.ItemBatonInitie, state.Equipment.RightHand.ItemCode)
	}

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	crafted, err := d2netpacket.UnmarshalItemCrafted(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal ItemCraftedPacket: %v", err)
	}

	if crafted.OutputItemCode != d2hero.ItemBatonInitie {
		t.Errorf("expected broadcast OutputItemCode %q, got %q", d2hero.ItemBatonInitie, crafted.OutputItemCode)
	}

	if crafted.OutputItemName != d2hero.DevilItems[d2hero.ItemBatonInitie].Name {
		t.Errorf("expected broadcast OutputItemName %q, got %q", d2hero.DevilItems[d2hero.ItemBatonInitie].Name, crafted.OutputItemName)
	}

	if crafted.Gold != state.Gold {
		t.Errorf("expected broadcast Gold %d, got %d", state.Gold, crafted.Gold)
	}
}

func TestResolveCraftWithoutRequiredItemIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Gold: 100}
	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateCraftRequestPacket("p", d2hero.RecipeUpgradeBatonApprenti)
	if err != nil {
		t.Fatalf("test setup: CreateCraftRequestPacket failed: %v", err)
	}

	server.resolveCraft(packet)

	if len(conn.sent) != 0 {
		t.Errorf("expected no packet sent for a failed craft, got %d", len(conn.sent))
	}
}

func TestResolveCraftUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateCraftRequestPacket("nobody", d2hero.RecipeUpgradeBatonApprenti)
	if err != nil {
		t.Fatalf("test setup: CreateCraftRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveCraft(packet)
}

func TestResolveLearnSkillLearnsAndBroadcastsRemainingPoints(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 1, SkillPoints: 2},
		Skills: make(map[int]*d2hero.HeroSkill),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateLearnSkillRequestPacket("p", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateLearnSkillRequestPacket failed: %v", err)
	}

	server.resolveLearnSkill(packet)

	if _, known := state.Skills[d2hero.SkillTraitDeFeu]; !known {
		t.Error("expected Trait de feu added to Skills")
	}

	if state.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints to drop to 1, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveLearnSkillFailsWithoutPointsIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 1, SkillPoints: 0},
		Skills: make(map[int]*d2hero.HeroSkill),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateLearnSkillRequestPacket("p", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateLearnSkillRequestPacket failed: %v", err)
	}

	server.resolveLearnSkill(packet)

	if _, known := state.Skills[d2hero.SkillTraitDeFeu]; known {
		t.Error("expected LearnSkill to fail with no skill points available")
	}
}

func TestResolveLearnSkillUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateLearnSkillRequestPacket("nobody", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateLearnSkillRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveLearnSkill(packet)
}

func TestResolveEquipSkillAssignsSlot(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 6},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: d2hero.NewDevilHeroSkill(d2hero.SkillTraitDeFeu)},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateEquipSkillRequestPacket("p", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateEquipSkillRequestPacket failed: %v", err)
	}

	server.resolveEquipSkill(packet)

	if state.LeftSkill != d2hero.SkillTraitDeFeu {
		t.Errorf("expected LeftSkill %d, got %d", d2hero.SkillTraitDeFeu, state.LeftSkill)
	}
}

func TestResolveEquipSkillFailsIfNotLearnedIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 6},
		Skills: make(map[int]*d2hero.HeroSkill),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateEquipSkillRequestPacket("p", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateEquipSkillRequestPacket failed: %v", err)
	}

	server.resolveEquipSkill(packet)

	if state.LeftSkill != 0 {
		t.Errorf("expected LeftSkill untouched at 0, got %d", state.LeftSkill)
	}
}

func TestResolveEquipSkillUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateEquipSkillRequestPacket("nobody", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateEquipSkillRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveEquipSkill(packet)
}

func TestResolveMoveToStashMovesTheItem(t *testing.T) {
	state := &d2hero.HeroState{
		Inventory: []string{d2hero.ItemPendentifArcane},
		Stash:     make([]string, 1),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveToStashRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToStashRequestPacket failed: %v", err)
	}

	server.resolveMoveToStash(packet)

	if state.Inventory[0] != "" {
		t.Errorf("expected the inventory slot cleared, got %q", state.Inventory[0])
	}

	if state.Stash[0] != d2hero.ItemPendentifArcane {
		t.Errorf("expected the item moved to the stash, got %q", state.Stash[0])
	}
}

func TestResolveMoveToStashFailsOnEmptySlotIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Inventory: make([]string, 1),
		Stash:     make([]string, 1),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveToStashRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToStashRequestPacket failed: %v", err)
	}

	server.resolveMoveToStash(packet)

	if state.Stash[0] != "" {
		t.Errorf("expected the stash untouched when the inventory slot was empty, got %q", state.Stash[0])
	}
}

func TestResolveMoveToStashUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateMoveToStashRequestPacket("nobody", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToStashRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveMoveToStash(packet)
}

func TestResolveMoveToInventoryMovesTheItem(t *testing.T) {
	state := &d2hero.HeroState{
		Inventory: make([]string, 1),
		Stash:     []string{d2hero.ItemAnneauDuDebut},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveToInventoryRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToInventoryRequestPacket failed: %v", err)
	}

	server.resolveMoveToInventory(packet)

	if state.Stash[0] != "" {
		t.Errorf("expected the stash slot cleared, got %q", state.Stash[0])
	}

	if state.Inventory[0] != d2hero.ItemAnneauDuDebut {
		t.Errorf("expected the item moved to the inventory, got %q", state.Inventory[0])
	}
}

func TestResolveMoveToInventoryUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateMoveToInventoryRequestPacket("nobody", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToInventoryRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveMoveToInventory(packet)
}

func TestResolveMoveToBeltMovesThePotion(t *testing.T) {
	state := &d2hero.HeroState{Inventory: []string{d2hero.ItemPotionDeMana}}
	state.InitBelt(1)
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveToBeltRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToBeltRequestPacket failed: %v", err)
	}

	server.resolveMoveToBelt(packet)

	if state.Inventory[0] != "" {
		t.Errorf("expected the inventory slot cleared, got %q", state.Inventory[0])
	}

	if state.Belt[0] != d2hero.ItemPotionDeMana {
		t.Errorf("expected the potion moved to the belt, got %q", state.Belt[0])
	}
}

func TestResolveMoveToBeltRejectsNonPotionIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Inventory: []string{d2hero.ItemPendentifArcane}}
	state.InitBelt(1)
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveToBeltRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToBeltRequestPacket failed: %v", err)
	}

	server.resolveMoveToBelt(packet)

	if state.Belt[0] != "" {
		t.Errorf("expected the belt untouched for a non-potion item, got %q", state.Belt[0])
	}
}

func TestResolveMoveToBeltUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateMoveToBeltRequestPacket("nobody", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveToBeltRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveMoveToBelt(packet)
}

func TestResolveMoveFromBeltMovesThePotion(t *testing.T) {
	state := &d2hero.HeroState{Inventory: make([]string, 1)}
	state.InitBelt(1)
	state.Belt[0] = d2hero.ItemPotionDeMana
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateMoveFromBeltRequestPacket("p", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveFromBeltRequestPacket failed: %v", err)
	}

	server.resolveMoveFromBelt(packet)

	if state.Belt[0] != "" {
		t.Errorf("expected the belt slot cleared, got %q", state.Belt[0])
	}

	if state.Inventory[0] != d2hero.ItemPotionDeMana {
		t.Errorf("expected the potion moved to the inventory, got %q", state.Inventory[0])
	}
}

func TestResolveMoveFromBeltUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateMoveFromBeltRequestPacket("nobody", 0)
	if err != nil {
		t.Fatalf("test setup: CreateMoveFromBeltRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveMoveFromBelt(packet)
}

func TestResolveRespecSkillsRefundsAllPoints(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: 1}},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSkillsRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateRespecSkillsRequestPacket failed: %v", err)
	}

	server.resolveRespecSkills(packet)

	if len(state.Skills) != 0 {
		t.Errorf("expected Skills cleared, got %d entries", len(state.Skills))
	}

	if state.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints refunded to 1, got %d", state.Stats.SkillPoints)
	}
}

// TestResolveRespecSkillsBroadcastsAttributesAndPools is a regression test:
// RespecSkills ("Respec partiel") also refunds every attribute point ever
// spent (HeroStatsState.RespecAllAttributePoints), but the broadcast
// packet used to carry only SkillPoints -- the client never learned its
// attributes/health/mana pools had reverted too.
func TestResolveRespecSkillsBroadcastsAttributesAndPools(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			Level: 6, SkillPoints: 0, StatsPoints: 1,
			Vitality: 10, Health: 40, MaxHealth: 40, LifePerVit: 4,
		},
		Skills: map[int]*d2hero.HeroSkill{},
	}

	if err := state.Stats.SpendAttributePoint(d2hero.AttributeVitality); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}
	// state.Stats is now Vitality=11, Health=MaxHealth=44, StatsPoints=0.

	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateRespecSkillsRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateRespecSkillsRequestPacket failed: %v", err)
	}

	server.resolveRespecSkills(packet)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	respeced, err := d2netpacket.UnmarshalSkillsRespeced(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal SkillsRespecedPacket: %v", err)
	}

	if respeced.Vitality != 10 {
		t.Errorf("expected the broadcast to carry Vitality reverted to 10, got %d", respeced.Vitality)
	}

	if respeced.MaxHealth != 40 || respeced.Health != 40 {
		t.Errorf("expected the broadcast to carry MaxHealth/Health reverted to 40/40, got %d/%d", respeced.MaxHealth, respeced.Health)
	}
}

func TestResolveRespecSkillsUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateRespecSkillsRequestPacket("nobody")
	if err != nil {
		t.Fatalf("test setup: CreateRespecSkillsRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveRespecSkills(packet)
}

// TestResolveRespecSkillsAlreadyUsedAtThisDifficultyIsNoop is a regression
// test for the design's "1 fois par difficulté" limit on Respec partiel.
func TestResolveRespecSkillsAlreadyUsedAtThisDifficultyIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:               &d2hero.HeroStatsState{SkillPoints: 0},
		Skills:              map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: 1}},
		RespecPartielUsedAt: map[d2enum.DifficultyType]bool{d2enum.DifficultyNormal: true},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSkillsRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateRespecSkillsRequestPacket failed: %v", err)
	}

	server.resolveRespecSkills(packet)

	if len(state.Skills) != 1 {
		t.Errorf("expected Skills untouched (already used this difficulty), got %d entries", len(state.Skills))
	}

	if state.Stats.SkillPoints != 0 {
		t.Errorf("expected no refund, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveRespecCompletConsumesEssenceAndRefunds(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:     &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills:    map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: 1}},
		Inventory: []string{d2hero.ItemEssenceDeBossOrdinaire},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecCompletRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateRespecCompletRequestPacket failed: %v", err)
	}

	server.resolveRespecComplet(packet)

	if len(state.Skills) != 0 {
		t.Errorf("expected Skills cleared, got %d entries", len(state.Skills))
	}

	if state.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints refunded to 1, got %d", state.Stats.SkillPoints)
	}

	if state.Inventory[0] != "" {
		t.Errorf("expected the essence consumed (slot cleared), got %q", state.Inventory[0])
	}
}

// TestResolveRespecCompletWithoutEssenceIsNoop is a regression test for the
// real point of this mechanism: without the item, nothing should happen --
// otherwise Respec complet would be freely available, defeating the whole
// rare-item gate it exists to implement.
func TestResolveRespecCompletWithoutEssenceIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:     &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills:    map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: 1}},
		Inventory: []string{""},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecCompletRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateRespecCompletRequestPacket failed: %v", err)
	}

	server.resolveRespecComplet(packet)

	if len(state.Skills) != 1 {
		t.Errorf("expected Skills untouched without the essence, got %d entries", len(state.Skills))
	}

	if state.Stats.SkillPoints != 0 {
		t.Errorf("expected no refund without the essence, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveRespecCompletUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateRespecCompletRequestPacket("nobody")
	if err != nil {
		t.Fatalf("test setup: CreateRespecCompletRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveRespecComplet(packet)
}

func TestResolveRespecSingleSkillRefundsOnlyThatSkill(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills: map[int]*d2hero.HeroSkill{
			d2hero.SkillTraitDeFeu:   {SkillPoints: 1},
			d2hero.SkillEclatDeGlace: {SkillPoints: 1},
		},
		Inventory: []string{d2hero.ItemGlypheDOubli},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleSkillRequestPacket("p", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleSkillRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleSkill(packet)

	if _, known := state.Skills[d2hero.SkillTraitDeFeu]; known {
		t.Error("expected Trait de feu forgotten")
	}

	if _, known := state.Skills[d2hero.SkillEclatDeGlace]; !known {
		t.Error("expected Éclat de glace to remain learned")
	}

	if state.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints refunded to 1, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveRespecSingleSkillFailsIfNotLearnedIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:     &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills:    make(map[int]*d2hero.HeroSkill),
		Inventory: []string{d2hero.ItemGlypheDOubli}, // present, so this actually exercises the "not learned" path, not "no glyph"
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleSkillRequestPacket("p", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleSkillRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleSkill(packet)

	if state.Stats.SkillPoints != 0 {
		t.Errorf("expected no refund for a skill that was never learned, got %d", state.Stats.SkillPoints)
	}
}

// TestResolveRespecSingleSkillWithoutGlyphIsNoop is the real point of
// UseGlypheDOubliOnSkill's gate: without the item, this must not forget
// the skill at all -- otherwise "Glyphe d'oubli" would be freely
// available, defeating the rare-item gate it exists to implement (same
// reasoning as TestResolveRespecCompletWithoutEssenceIsNoop).
func TestResolveRespecSingleSkillWithoutGlyphIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 6, SkillPoints: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: 1}},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleSkillRequestPacket("p", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleSkillRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleSkill(packet)

	if _, known := state.Skills[d2hero.SkillTraitDeFeu]; !known {
		t.Error("expected Trait de feu untouched without the glyph in inventory")
	}

	if state.Stats.SkillPoints != 0 {
		t.Errorf("expected no refund without the glyph, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveRespecSingleSkillUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateRespecSingleSkillRequestPacket("nobody", d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleSkillRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveRespecSingleSkill(packet)
}

func TestResolveInvestSkillPointAddsPointAndBroadcastsTotals(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 18, SkillPoints: 2},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillMaitriseElementaire: {SkillPoints: 1}},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateInvestSkillPointRequestPacket("p", d2hero.SkillMaitriseElementaire)
	if err != nil {
		t.Fatalf("test setup: CreateInvestSkillPointRequestPacket failed: %v", err)
	}

	server.resolveInvestSkillPoint(packet)

	if got := state.Skills[d2hero.SkillMaitriseElementaire].SkillPoints; got != 2 {
		t.Errorf("expected 2 points invested, got %d", got)
	}

	if state.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints to drop to 1, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveInvestSkillPointFailsIfNotLearnedIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Level: 18, SkillPoints: 2},
		Skills: make(map[int]*d2hero.HeroSkill),
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateInvestSkillPointRequestPacket("p", d2hero.SkillMaitriseElementaire)
	if err != nil {
		t.Fatalf("test setup: CreateInvestSkillPointRequestPacket failed: %v", err)
	}

	server.resolveInvestSkillPoint(packet)

	if state.Stats.SkillPoints != 2 {
		t.Errorf("expected no point spent investing in an unlearned skill, got %d", state.Stats.SkillPoints)
	}
}

func TestResolveInvestSkillPointUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateInvestSkillPointRequestPacket("nobody", d2hero.SkillMaitriseElementaire)
	if err != nil {
		t.Fatalf("test setup: CreateInvestSkillPointRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveInvestSkillPoint(packet)
}

func TestResolveSpendAttributePointSpendsAndBroadcastsNewValue(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{StatsPoints: 2, Vitality: 10}}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateSpendAttributePointRequestPacket("p", int(d2hero.AttributeVitality))
	if err != nil {
		t.Fatalf("test setup: CreateSpendAttributePointRequestPacket failed: %v", err)
	}

	server.resolveSpendAttributePoint(packet)

	if state.Stats.Vitality != 11 {
		t.Errorf("expected Vitality=11, got %d", state.Stats.Vitality)
	}

	if state.Stats.StatsPoints != 1 {
		t.Errorf("expected StatsPoints to drop to 1, got %d", state.Stats.StatsPoints)
	}
}

// TestResolveSpendAttributePointBroadcastsGrownHealth is a regression test:
// spending on Vitality also grows MaxHealth/Health (HeroStatsState.
// SpendAttributePoint), but the broadcast packet used to carry only the
// attribute itself -- the client's own health pool never learned it grew.
func TestResolveSpendAttributePointBroadcastsGrownHealth(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{StatsPoints: 1, Vitality: 10, Health: 40, MaxHealth: 40, LifePerVit: 4},
	}
	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateSpendAttributePointRequestPacket("p", int(d2hero.AttributeVitality))
	if err != nil {
		t.Fatalf("test setup: CreateSpendAttributePointRequestPacket failed: %v", err)
	}

	server.resolveSpendAttributePoint(packet)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	spent, err := d2netpacket.UnmarshalAttributePointSpent(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal AttributePointSpentPacket: %v", err)
	}

	if spent.MaxHealth != 44 || spent.Health != 44 {
		t.Errorf("expected the broadcast to carry the grown MaxHealth/Health (44/44), got %d/%d", spent.MaxHealth, spent.Health)
	}
}

func TestResolveSpendAttributePointFailsWithoutPointsIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{StatsPoints: 0, Vitality: 10}}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateSpendAttributePointRequestPacket("p", int(d2hero.AttributeVitality))
	if err != nil {
		t.Fatalf("test setup: CreateSpendAttributePointRequestPacket failed: %v", err)
	}

	server.resolveSpendAttributePoint(packet)

	if state.Stats.Vitality != 10 {
		t.Errorf("expected Vitality unchanged at 10, got %d", state.Stats.Vitality)
	}
}

func TestResolveSpendAttributePointUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateSpendAttributePointRequestPacket("nobody", int(d2hero.AttributeVitality))
	if err != nil {
		t.Fatalf("test setup: CreateSpendAttributePointRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveSpendAttributePoint(packet)
}

func TestResolveRespecSingleAttributePointRefundsOnlyThatAttribute(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:     &d2hero.HeroStatsState{StatsPoints: 0, Strength: 11, StrengthSpent: 1},
		Inventory: []string{d2hero.ItemGlypheDOubli},
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleAttributePointRequestPacket("p", int(d2hero.AttributeStrength))
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleAttributePointRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleAttributePoint(packet)

	if state.Stats.Strength != 10 {
		t.Errorf("expected Strength restored to 10, got %d", state.Stats.Strength)
	}

	if state.Stats.StatsPoints != 1 {
		t.Errorf("expected the point refunded to StatsPoints, got %d", state.Stats.StatsPoints)
	}
}

// TestResolveRespecSingleAttributePointBroadcastsShrunkHealth is a
// regression test: refunding a Vitality point also shrinks MaxHealth/
// Health (HeroStatsState.RefundAttributePoint), but the broadcast packet
// used to carry only the attribute itself.
func TestResolveRespecSingleAttributePointBroadcastsShrunkHealth(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			StatsPoints: 0, Vitality: 11, VitalitySpent: 1, LifePerVit: 4, Health: 44, MaxHealth: 44,
		},
		Inventory: []string{d2hero.ItemGlypheDOubli},
	}
	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateRespecSingleAttributePointRequestPacket("p", int(d2hero.AttributeVitality))
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleAttributePointRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleAttributePoint(packet)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	respeced, err := d2netpacket.UnmarshalSingleAttributePointRespeced(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal SingleAttributePointRespecedPacket: %v", err)
	}

	if respeced.MaxHealth != 40 || respeced.Health != 40 {
		t.Errorf("expected the broadcast to carry the shrunk MaxHealth/Health (40/40), got %d/%d", respeced.MaxHealth, respeced.Health)
	}
}

func TestResolveRespecSingleAttributePointFailsIfNothingSpentIsNoop(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:     &d2hero.HeroStatsState{StatsPoints: 0, Strength: 10},
		Inventory: []string{d2hero.ItemGlypheDOubli}, // present, so this actually exercises the "nothing spent" path, not "no glyph"
	}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleAttributePointRequestPacket("p", int(d2hero.AttributeStrength))
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleAttributePointRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleAttributePoint(packet)

	if state.Stats.StatsPoints != 0 {
		t.Errorf("expected no refund with nothing spent, got %d", state.Stats.StatsPoints)
	}
}

// TestResolveRespecSingleAttributePointWithoutGlyphIsNoop mirrors
// TestResolveRespecSingleSkillWithoutGlyphIsNoop: without the item, this
// must not refund the point at all, or "Glyphe d'oubli" would be freely
// available, defeating the rare-item gate it exists to implement.
func TestResolveRespecSingleAttributePointWithoutGlyphIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{StatsPoints: 0, Strength: 11, StrengthSpent: 1}}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateRespecSingleAttributePointRequestPacket("p", int(d2hero.AttributeStrength))
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleAttributePointRequestPacket failed: %v", err)
	}

	server.resolveRespecSingleAttributePoint(packet)

	if state.Stats.Strength != 11 || state.Stats.StrengthSpent != 1 {
		t.Errorf("expected Strength untouched without the glyph, got Strength=%d StrengthSpent=%d",
			state.Stats.Strength, state.Stats.StrengthSpent)
	}

	if state.Stats.StatsPoints != 0 {
		t.Errorf("expected no refund without the glyph, got %d", state.Stats.StatsPoints)
	}
}

func TestResolveRespecSingleAttributePointUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateRespecSingleAttributePointRequestPacket("nobody", int(d2hero.AttributeStrength))
	if err != nil {
		t.Fatalf("test setup: CreateRespecSingleAttributePointRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveRespecSingleAttributePoint(packet)
}

func TestRestoreManaOnKillWithAbsorptionEnergieLearned(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Mana: 0, MaxMana: 20},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillAbsorptionEnergie: {}},
	}
	server := serverWithConnection(state)

	server.restoreManaOnKill("p")

	if want := 20 * absorptionEnergieManaRestorePercent / 100; state.Stats.Mana != want {
		t.Errorf("expected Mana restored to %d, got %d", want, state.Stats.Mana)
	}
}

// TestRestoreManaOnKillBroadcastsNewMana is a regression test: the restore
// used to mutate the server's own copy of Mana with nothing telling the
// client (even in solo play, which still round-trips through a local
// client/server -- ROADMAP.md).
func TestRestoreManaOnKillBroadcastsNewMana(t *testing.T) {
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Mana: 0, MaxMana: 20},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillAbsorptionEnergie: {}},
	}
	conn := &fakeClientConnection{state: state}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	server.restoreManaOnKill("p")

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	used, err := d2netpacket.UnmarshalPotionUsed(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PotionUsedPacket: %v", err)
	}

	if used.Mana != state.Stats.Mana {
		t.Errorf("expected the broadcast to carry the restored Mana (%d), got %d", state.Stats.Mana, used.Mana)
	}
}

func TestRestoreManaOnKillWithoutAbsorptionEnergieIsNoop(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Mana: 0, MaxMana: 20}}
	server := serverWithConnection(state)

	server.restoreManaOnKill("p")

	if state.Stats.Mana != 0 {
		t.Errorf("expected no mana restored without the skill learned, got %d", state.Stats.Mana)
	}
}

func TestRestoreManaOnKillUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	// must not panic when the player isn't a connected/resolved state.
	server.restoreManaOnKill("nobody")
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

func TestEffectiveEnergyIncludesEquippedWeaponBonus(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: 20},
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	}

	if got, want := effectiveEnergy(state), 20+d2hero.ItemEnergyBonus(d2hero.ItemBatonApprenti); got != want {
		t.Errorf("expected effective Energy %d, got %d", want, got)
	}
}

func TestEffectiveEnergyWithNoWeaponIsJustBaseEnergy(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 20}}

	if got := effectiveEnergy(state); got != 20 {
		t.Errorf("expected effective Energy 20 with nothing equipped, got %d", got)
	}
}

func TestEffectiveEnergyStacksWeaponAndAmuletBonuses(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: 20},
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
			Amulet:    &d2inventory.InventoryItemMisc{ItemCode: d2hero.ItemPendentifArcane},
		},
	}

	want := 20 + d2hero.ItemEnergyBonus(d2hero.ItemBatonApprenti) + d2hero.ItemEnergyBonus(d2hero.ItemPendentifArcane)

	if got := effectiveEnergy(state); got != want {
		t.Errorf("expected effective Energy %d (base + both equipped bonuses), got %d", want, got)
	}
}

func TestEffectiveEnergyIncludesRingAllAttributesBonus(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: 20},
		Equipment: d2inventory.CharacterEquipment{
			Ring: &d2inventory.InventoryItemMisc{ItemCode: d2hero.ItemAnneauDuDebut},
		},
	}

	want := 20 + d2hero.ItemAllAttributesBonus(d2hero.ItemAnneauDuDebut)

	if got := effectiveEnergy(state); got != want {
		t.Errorf("expected effective Energy %d (base + ring's all-attributes bonus), got %d", want, got)
	}
}

func TestEffectiveDexterityIncludesRingAllAttributesBonus(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Dexterity: 15},
		Equipment: d2inventory.CharacterEquipment{
			Ring: &d2inventory.InventoryItemMisc{ItemCode: d2hero.ItemAnneauDuDebut},
		},
	}

	want := 15 + d2hero.ItemAllAttributesBonus(d2hero.ItemAnneauDuDebut)

	if got := effectiveDexterity(state); got != want {
		t.Errorf("expected effective Dexterity %d (base + ring's all-attributes bonus), got %d", want, got)
	}
}

func TestEffectiveDexterityWithNothingEquippedIsJustBaseDexterity(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{Dexterity: 15}}

	if got := effectiveDexterity(state); got != 15 {
		t.Errorf("expected effective Dexterity 15 with nothing equipped, got %d", got)
	}
}

// TestResolveAttackDamageIncludesEquippedWeaponEnergyBonus is a regression
// test: DevilItemDef.EnergyBonus (e.g. Bâton de l'Apprenti's declared
// "+5 Energy", devil_mage_character_design.md §5) previously had zero
// effect on gameplay -- resolveAttackDamage read state.Stats.Energy
// directly, never consulting an equipped item's bonus.
func TestResolveAttackDamageIncludesEquippedWeaponEnergyBonus(t *testing.T) {
	const energy = 13 // chosen so the item's Energy bonus visibly changes
	// the truncated damage total below -- see the sanity check.

	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: energy},
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	})

	traitDeFeuBase := d2hero.DevilSkills[d2hero.SkillTraitDeFeu].BaseSortDamage
	withoutBonus := traitDeFeuBase + (traitDeFeuBase*energy)/100
	withBonus := traitDeFeuBase + (traitDeFeuBase*(energy+d2hero.ItemEnergyBonus(d2hero.ItemBatonApprenti)))/100

	if withBonus <= withoutBonus {
		t.Fatal("test setup error: chosen Energy doesn't make the item's bonus visible after integer truncation")
	}

	want := withBonus + (withBonus*d2hero.ItemFireDamagePercent(d2hero.ItemBatonApprenti))/100

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != want {
		t.Errorf("expected %d (Energy bonus + Fire modifier both applied), got %d", want, got)
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

// TestResolveAttackDamageWeaponFireModifierDoesNotLeakToNonFireSkills is a
// regression test: a Fire%-boosting weapon used to buff every skill's
// damage regardless of its actual element, since resolveAttackDamage had
// no per-skill element data to check against.
func TestResolveAttackDamageWeaponFireModifierDoesNotLeakToNonFireSkills(t *testing.T) {
	const eclatDeGlaceBase = 4 // no Energy scaling in this test (Energy: 0)

	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{Energy: 0},
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	})

	if got := server.resolveAttackDamage("p", d2hero.SkillEclatDeGlace); got != eclatDeGlaceBase {
		t.Errorf("expected no Fire modifier on a Cold skill, got %d instead of the base %d", got, eclatDeGlaceBase)
	}
}

func TestResolveAttackDamageNoWeaponNoModifier(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{Stats: &d2hero.HeroStatsState{Energy: 0}})

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != 6 {
		t.Errorf("expected no modifier with nothing equipped, got %d", got)
	}
}

// TestResolveAttackDamageAppliesMaitriseElementaireToElementalisteSkills is a
// regression test for Maîtrise élémentaire ("+% dégâts élémentaires par
// point", devil_game_design_reference.md §7) -- Devil's first
// multi-point-investment skill (see HeroState.InvestSkillPoint).
func TestResolveAttackDamageAppliesMaitriseElementaireToElementalisteSkills(t *testing.T) {
	const points = 3

	server := serverWithConnection(&d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Energy: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillMaitriseElementaire: {SkillPoints: points}},
	})

	traitDeFeuBase := d2hero.DevilSkills[d2hero.SkillTraitDeFeu].BaseSortDamage
	percent := d2hero.MaitriseElementaireDamagePercent(points)
	want := traitDeFeuBase + (traitDeFeuBase*percent)/100

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != want {
		t.Errorf("expected Maîtrise élémentaire's bonus applied to an Élémentalisme skill (%d), got %d", want, got)
	}
}

// TestResolveAttackDamageMaitriseElementaireDoesNotAffectOtherTrees checks
// the tree gating: Maîtrise élémentaire must not inflate an Ésotérisme
// skill's damage.
func TestResolveAttackDamageMaitriseElementaireDoesNotAffectOtherTrees(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Energy: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillMaitriseElementaire: {SkillPoints: 3}},
	})

	want := d2hero.DevilSkills[d2hero.SkillTempeteDeLames].BaseSortDamage

	if got := server.resolveAttackDamage("p", d2hero.SkillTempeteDeLames); got != want {
		t.Errorf("expected an Ésotérisme skill to be unaffected by Maîtrise élémentaire (%d), got %d", want, got)
	}
}

// TestResolveAttackDamageAppliesTraitDeFeuSynergyToItsTargetsOnly is a
// regression test for Trait de feu's own synergy ("chaque point dans Trait
// de feu augmente les dégâts de Boule de feu et Météore",
// devil_game_design_reference.md §7) -- narrower than Maîtrise élémentaire,
// which boosts every Élémentalisme skill: this only targets Boule de
// feu/Météore specifically.
func TestResolveAttackDamageAppliesTraitDeFeuSynergyToItsTargetsOnly(t *testing.T) {
	const points = 4

	server := serverWithConnection(&d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{Energy: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {SkillPoints: points}},
	})

	percent := d2hero.TraitDeFeuSynergyDamagePercent(points)

	bouleDeFeuBase := d2hero.DevilSkills[d2hero.SkillBouleDeFeu].BaseSortDamage
	wantBouleDeFeu := bouleDeFeuBase + (bouleDeFeuBase*percent)/100

	if got := server.resolveAttackDamage("p", d2hero.SkillBouleDeFeu); got != wantBouleDeFeu {
		t.Errorf("expected Trait de feu's synergy applied to Boule de feu (%d), got %d", wantBouleDeFeu, got)
	}

	meteoreBase := d2hero.DevilSkills[d2hero.SkillMeteore].BaseSortDamage
	wantMeteore := meteoreBase + (meteoreBase*percent)/100

	if got := server.resolveAttackDamage("p", d2hero.SkillMeteore); got != wantMeteore {
		t.Errorf("expected Trait de feu's synergy applied to Météore (%d), got %d", wantMeteore, got)
	}

	// A third Élémentalisme skill, not one of the synergy's two named
	// targets, must be unaffected -- unlike Maîtrise élémentaire's tree-wide
	// gating, this synergy only targets Boule de feu/Météore specifically.
	eclatDeGlaceBase := d2hero.DevilSkills[d2hero.SkillEclatDeGlace].BaseSortDamage

	if got := server.resolveAttackDamage("p", d2hero.SkillEclatDeGlace); got != eclatDeGlaceBase {
		t.Errorf("expected Éclat de glace unaffected by Trait de feu's synergy (%d), got %d", eclatDeGlaceBase, got)
	}
}

// TestResolveAttackDamageConsumesResonanceMagiqueBonus is a regression test
// for Résonance magique ("chaque sort lancé augmente les dégâts du
// suivant", devil_game_design_reference.md §7). The bonus itself is armed
// by resolveMeleeHit on every cast (untestable here -- same map-engine/real
// NPC gap as the rest of that function, see ROADMAP.md); this exercises
// resolveAttackDamage's own consumption of an already-armed bonus.
func TestResolveAttackDamageConsumesResonanceMagiqueBonus(t *testing.T) {
	const points = 2

	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			Energy:                     0,
			ResonanceMagiqueBonusUntil: time.Now().Add(time.Minute),
		},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillResonanceMagique: {SkillPoints: points}},
	})

	traitDeFeuBase := d2hero.DevilSkills[d2hero.SkillTraitDeFeu].BaseSortDamage
	percent := d2hero.ResonanceMagiqueDamagePercent(points)
	want := traitDeFeuBase + (traitDeFeuBase*percent)/100

	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != want {
		t.Errorf("expected Résonance magique's armed bonus applied (%d), got %d", want, got)
	}
}

// TestResolveAttackDamageResonanceMagiqueConsumedOnce checks that a second
// cast right after the first no longer benefits -- the bonus is meant for
// exactly one "next" cast.
func TestResolveAttackDamageResonanceMagiqueConsumedOnce(t *testing.T) {
	state := &d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			Energy:                     0,
			ResonanceMagiqueBonusUntil: time.Now().Add(time.Minute),
		},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillResonanceMagique: {SkillPoints: 2}},
	}
	server := serverWithConnection(state)

	server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu)

	traitDeFeuBase := d2hero.DevilSkills[d2hero.SkillTraitDeFeu].BaseSortDamage
	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != traitDeFeuBase {
		t.Errorf("expected the bonus already spent by the first cast, got %d instead of the base %d", got, traitDeFeuBase)
	}
}

// TestResolveAttackDamageResonanceMagiqueRequiresTheSkillLearned checks that
// an armed-but-unconsumed bonus does nothing for a caster who never learned
// Résonance magique.
func TestResolveAttackDamageResonanceMagiqueRequiresTheSkillLearned(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			Energy:                     0,
			ResonanceMagiqueBonusUntil: time.Now().Add(time.Minute),
		},
		Skills: map[int]*d2hero.HeroSkill{},
	})

	traitDeFeuBase := d2hero.DevilSkills[d2hero.SkillTraitDeFeu].BaseSortDamage
	if got := server.resolveAttackDamage("p", d2hero.SkillTraitDeFeu); got != traitDeFeuBase {
		t.Errorf("expected no bonus without the skill learned, got %d instead of the base %d", got, traitDeFeuBase)
	}
}

// TestResolveAttackDamageResonanceMagiqueDoesNotAffectEsoterisme checks the
// tree gating: Résonance magique reaches Élémentalisme/Arcane, not
// Ésotérisme.
func TestResolveAttackDamageResonanceMagiqueDoesNotAffectEsoterisme(t *testing.T) {
	server := serverWithConnection(&d2hero.HeroState{
		Stats: &d2hero.HeroStatsState{
			Energy:                     0,
			ResonanceMagiqueBonusUntil: time.Now().Add(time.Minute),
		},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillResonanceMagique: {SkillPoints: 3}},
	})

	want := d2hero.DevilSkills[d2hero.SkillTempeteDeLames].BaseSortDamage

	if got := server.resolveAttackDamage("p", d2hero.SkillTempeteDeLames); got != want {
		t.Errorf("expected an Ésotérisme skill unaffected by Résonance magique (%d), got %d", want, got)
	}
}

// TestResonanceMagiqueBonusPercentReachesChampStatique is a regression
// test: Champ statique deals a % of the target's current HP via its own
// resolveChampStatiqueHit, bypassing resolveAttackDamage entirely -- so it
// never consumed Résonance magique's armed bonus even though it's squarely
// an Arcane "sort actif" the synergy is meant to reach
// (devil_game_design_reference.md §7 "Synergies": "tous les sorts actifs
// des arbres I et II"). resonanceMagiqueBonusPercent is the helper both
// paths now share.
func TestResonanceMagiqueBonusPercentReachesChampStatique(t *testing.T) {
	const points = 2

	server := serverWithConnection(nil)
	state := &d2hero.HeroState{
		Stats:  &d2hero.HeroStatsState{ResonanceMagiqueBonusUntil: time.Now().Add(time.Minute)},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillResonanceMagique: {SkillPoints: points}},
	}

	want := d2hero.ResonanceMagiqueDamagePercent(points)

	if got := server.resonanceMagiqueBonusPercent(state, d2hero.SkillChampStatique); got != want {
		t.Errorf("expected Champ statique to consume the armed bonus (%d), got %d", want, got)
	}
}

func TestResolveToggleOverloadTogglesOverloadActive(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateToggleOverloadRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateToggleOverloadRequestPacket failed: %v", err)
	}

	if state.Stats.OverloadActive {
		t.Fatal("expected OverloadActive to start false")
	}

	server.resolveToggleOverload(packet)

	if !state.Stats.OverloadActive {
		t.Error("expected first toggle to turn OverloadActive on")
	}

	server.resolveToggleOverload(packet)

	if state.Stats.OverloadActive {
		t.Error("expected second toggle to turn OverloadActive back off")
	}
}

func TestResolveToggleOverloadUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateToggleOverloadRequestPacket("nobody")
	if err != nil {
		t.Fatalf("test setup: CreateToggleOverloadRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveToggleOverload(packet)
}

func TestResolveToggleOverloadBroadcastsToggle(t *testing.T) {
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateToggleOverloadRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateToggleOverloadRequestPacket failed: %v", err)
	}

	server.resolveToggleOverload(packet)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalPlayerStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.PlayerStatusOverload || !status.Active {
		t.Errorf("expected effect %q active=true, got %q active=%v", d2netpacket.PlayerStatusOverload, status.Effect, status.Active)
	}
}

func TestResolveToggleMarqueArdenteTogglesMarqueArdenteActive(t *testing.T) {
	state := &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}
	server := serverWithConnection(state)

	packet, err := d2netpacket.CreateToggleMarqueArdenteRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateToggleMarqueArdenteRequestPacket failed: %v", err)
	}

	if state.Stats.MarqueArdenteActive {
		t.Fatal("expected MarqueArdenteActive to start false")
	}

	server.resolveToggleMarqueArdente(packet)

	if !state.Stats.MarqueArdenteActive {
		t.Error("expected first toggle to turn MarqueArdenteActive on")
	}

	server.resolveToggleMarqueArdente(packet)

	if state.Stats.MarqueArdenteActive {
		t.Error("expected second toggle to turn MarqueArdenteActive back off")
	}
}

func TestResolveToggleMarqueArdenteUnknownPlayerNoop(t *testing.T) {
	server := serverWithConnection(nil)

	packet, err := d2netpacket.CreateToggleMarqueArdenteRequestPacket("nobody")
	if err != nil {
		t.Fatalf("test setup: CreateToggleMarqueArdenteRequestPacket failed: %v", err)
	}

	// must not panic when the caster isn't a connected/resolved player.
	server.resolveToggleMarqueArdente(packet)
}

func TestResolveToggleMarqueArdenteBroadcastsToggle(t *testing.T) {
	conn := &fakeClientConnection{state: &d2hero.HeroState{Stats: &d2hero.HeroStatsState{}}}
	server := &GameServer{connections: map[string]ClientConnection{"p": conn}}

	packet, err := d2netpacket.CreateToggleMarqueArdenteRequestPacket("p")
	if err != nil {
		t.Fatalf("test setup: CreateToggleMarqueArdenteRequestPacket failed: %v", err)
	}

	server.resolveToggleMarqueArdente(packet)

	if len(conn.sent) != 1 {
		t.Fatalf("expected 1 packet sent, got %d", len(conn.sent))
	}

	status, err := d2netpacket.UnmarshalPlayerStatusEffect(conn.sent[0].PacketData)
	if err != nil {
		t.Fatalf("failed to unmarshal PlayerStatusEffectPacket: %v", err)
	}

	if status.Effect != d2netpacket.PlayerStatusMarqueArdente || !status.Active {
		t.Errorf("expected effect %q active=true, got %q active=%v",
			d2netpacket.PlayerStatusMarqueArdente, status.Effect, status.Active)
	}
}

// TestCanCastNowChargesOverloadSurcharge is a regression test for
// Overload's mana surcharge actually being charged, not just declared:
// without OverloadActive gating the cost, this would silently charge the
// skill's normal cost instead.
func TestCanCastNowChargesOverloadSurcharge(t *testing.T) {
	stats := &d2hero.HeroStatsState{Mana: 10, MaxMana: 10, OverloadActive: true}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	// Trait de feu costs 3; with Overload's placeholder +50% surcharge
	// that's 4 (integer division), not 3.
	if !server.canCastNow("p", d2hero.SkillTraitDeFeu) {
		t.Fatal("expected the cast to succeed with enough mana")
	}

	wantCost := 3 * overloadManaCostPercent / 100
	if got := 10 - stats.Mana; got != wantCost {
		t.Errorf("expected the surcharged cost (%d) deducted, got %d", wantCost, got)
	}
}

func TestCanCastNowWithoutOverloadChargesNormalCost(t *testing.T) {
	stats := &d2hero.HeroStatsState{Mana: 10, MaxMana: 10, OverloadActive: false}
	server := serverWithConnection(&d2hero.HeroState{Stats: stats})

	server.canCastNow("p", d2hero.SkillTraitDeFeu)

	if stats.Mana != 7 {
		t.Errorf("expected the normal cost (3) deducted without Overload, got mana %d", stats.Mana)
	}
}
