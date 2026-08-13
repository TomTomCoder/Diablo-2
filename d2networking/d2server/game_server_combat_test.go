package d2server

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
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

func serverWithWeapons(t *testing.T, weapons d2records.CommonItems) *GameServer {
	t.Helper()

	assetManager, err := d2asset.NewAssetManager(d2util.LogLevelDefault)
	if err != nil {
		t.Fatalf("NewAssetManager: %v", err)
	}

	assetManager.Records.Item.Weapons = weapons

	return &GameServer{asset: assetManager, connections: make(map[string]ClientConnection)}
}

func TestResolveAttackDamageUnarmedFallbacks(t *testing.T) {
	server := serverWithWeapons(t, d2records.CommonItems{})

	cases := map[string]*d2hero.HeroState{
		"no connection at all": nil, // handled by not registering a connection
		"empty equipment":      {Equipment: d2inventory.CharacterEquipment{}},
		"weapon code not in records": {Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: "nonexistent"},
		}},
	}

	for name, state := range cases {
		if state != nil {
			server.connections["p"] = &fakeClientConnection{state: state}
		}

		minDmg, maxDmg := server.resolveAttackDamage("p")
		if minDmg != unarmedMinDamage || maxDmg != unarmedMaxDamage {
			t.Errorf("%s: expected unarmed damage (%d,%d), got (%d,%d)",
				name, unarmedMinDamage, unarmedMaxDamage, minDmg, maxDmg)
		}
	}
}

func TestResolveAttackDamageOneHanded(t *testing.T) {
	server := serverWithWeapons(t, d2records.CommonItems{
		"shortsword": {MinDamage: 3, MaxDamage: 7},
	})
	server.connections["p"] = &fakeClientConnection{state: &d2hero.HeroState{
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: "shortsword"},
		},
	}}

	minDmg, maxDmg := server.resolveAttackDamage("p")
	if minDmg != 3 || maxDmg != 7 {
		t.Errorf("expected (3,7), got (%d,%d)", minDmg, maxDmg)
	}
}

func TestResolveAttackDamageTwoHandedFallback(t *testing.T) {
	// A weapon with no one-handed range should fall back to its two-handed
	// range rather than reporting (0,0).
	server := serverWithWeapons(t, d2records.CommonItems{
		"greatsword": {Min2HandDamage: 10, Max2HandDamage: 20},
	})
	server.connections["p"] = &fakeClientConnection{state: &d2hero.HeroState{
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: "greatsword"},
		},
	}}

	minDmg, maxDmg := server.resolveAttackDamage("p")
	if minDmg != 10 || maxDmg != 20 {
		t.Errorf("expected two-hand fallback (10,20), got (%d,%d)", minDmg, maxDmg)
	}
}

func TestResolveAttackDamageLeftHandFallback(t *testing.T) {
	// Empty right hand should fall back to the left hand weapon.
	server := serverWithWeapons(t, d2records.CommonItems{
		"dagger": {MinDamage: 1, MaxDamage: 4},
	})
	server.connections["p"] = &fakeClientConnection{state: &d2hero.HeroState{
		Equipment: d2inventory.CharacterEquipment{
			LeftHand: &d2inventory.InventoryItemWeapon{ItemCode: "dagger"},
		},
	}}

	minDmg, maxDmg := server.resolveAttackDamage("p")
	if minDmg != 1 || maxDmg != 4 {
		t.Errorf("expected left-hand weapon (1,4), got (%d,%d)", minDmg, maxDmg)
	}
}
