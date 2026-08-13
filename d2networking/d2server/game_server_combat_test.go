package d2server

import (
	"testing"

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
	server := &GameServer{connections: make(map[string]ClientConnection)}
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
