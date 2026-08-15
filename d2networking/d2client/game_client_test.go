package d2client

import (
	"testing"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket"
)

// This is the first test file for package d2client: every handler here
// only ever touches GameClient.Players (a plain exported map), so each is
// unit-testable without a real map engine or asset manager. The NPC-facing
// handlers (handleNPCMovedPacket/handleNPCStatusEffectPacket) aren't
// covered here -- they need a populated MapEngine.Entities(), which is
// only ever initialized by ResetMap against a real, MPQ-loaded asset
// manager (see d2mapengine.MapEngine.ResetMap), the same limitation
// documented throughout d2server's own tests.

func clientWithPlayer(id string, player *d2mapentity.Player) *GameClient {
	return &GameClient{Players: map[string]*d2mapentity.Player{id: player}}
}

func TestHandlePlayerStatusEffectPacketManaShield(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{}})

	packet, err := d2netpacket.CreatePlayerStatusEffectPacket("p", d2netpacket.PlayerStatusManaShield, true, time.Time{})
	if err != nil {
		t.Fatalf("test setup: CreatePlayerStatusEffectPacket failed: %v", err)
	}

	if err := client.handlePlayerStatusEffectPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !client.Players["p"].Stats.ManaShieldActive {
		t.Error("expected ManaShieldActive to be applied")
	}
}

func TestHandlePlayerStatusEffectPacketOverload(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{}})

	packet, err := d2netpacket.CreatePlayerStatusEffectPacket("p", d2netpacket.PlayerStatusOverload, true, time.Time{})
	if err != nil {
		t.Fatalf("test setup: CreatePlayerStatusEffectPacket failed: %v", err)
	}

	if err := client.handlePlayerStatusEffectPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !client.Players["p"].Stats.OverloadActive {
		t.Error("expected OverloadActive to be applied")
	}
}

func TestHandlePlayerStatusEffectPacketMagicImmune(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{}})

	until := time.Now().Add(8 * time.Second)

	packet, err := d2netpacket.CreatePlayerStatusEffectPacket("p", d2netpacket.PlayerStatusMagicImmune, true, until)
	if err != nil {
		t.Fatalf("test setup: CreatePlayerStatusEffectPacket failed: %v", err)
	}

	if err := client.handlePlayerStatusEffectPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !client.Players["p"].Stats.IsMagicImmune(time.Now()) {
		t.Error("expected magic immunity to be applied")
	}
}

func TestHandleAttributePointSpentPacketAppliesAttributeAndPools(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{Vitality: 10, MaxHealth: 40, Health: 40}})

	packet, err := d2netpacket.CreateAttributePointSpentPacket("p", int(d2hero.AttributeVitality), 11, 0, 44, 44, 20, 20)
	if err != nil {
		t.Fatalf("test setup: CreateAttributePointSpentPacket failed: %v", err)
	}

	if err := client.handleAttributePointSpentPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stats := client.Players["p"].Stats
	if stats.Vitality != 11 || stats.MaxHealth != 44 || stats.Health != 44 {
		t.Errorf("expected Vitality=11 MaxHealth=Health=44, got Vitality=%d MaxHealth=%d Health=%d",
			stats.Vitality, stats.MaxHealth, stats.Health)
	}
}

func TestHandleSingleAttributePointRespecedPacketAppliesAttributeAndPools(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{Vitality: 11, MaxHealth: 44, Health: 44}})

	packet, err := d2netpacket.CreateSingleAttributePointRespecedPacket("p", int(d2hero.AttributeVitality), 10, 1, 40, 40, 20, 20)
	if err != nil {
		t.Fatalf("test setup: CreateSingleAttributePointRespecedPacket failed: %v", err)
	}

	if err := client.handleSingleAttributePointRespecedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stats := client.Players["p"].Stats
	if stats.Vitality != 10 || stats.MaxHealth != 40 || stats.Health != 40 || stats.StatsPoints != 1 {
		t.Errorf("expected Vitality=10 MaxHealth=Health=40 StatsPoints=1, got Vitality=%d MaxHealth=%d Health=%d StatsPoints=%d",
			stats.Vitality, stats.MaxHealth, stats.Health, stats.StatsPoints)
	}
}

func TestHandleSkillsRespecedPacketAppliesEverything(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Stats:      &d2hero.HeroStatsState{},
		Skills:     map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {}},
		LeftSkill:  d2hero.NewTraitDeFeuSkill(),
		RightSkill: d2hero.NewTraitDeFeuSkill(),
	})

	packet, err := d2netpacket.CreateSkillsRespecedPacket("p", 2, 3, 10, 10, 10, 10, 40, 40, 20, 20)
	if err != nil {
		t.Fatalf("test setup: CreateSkillsRespecedPacket failed: %v", err)
	}

	if err := client.handleSkillsRespecedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if len(player.Skills) != 0 {
		t.Errorf("expected Skills cleared, got %d entries", len(player.Skills))
	}

	if player.LeftSkill != nil || player.RightSkill != nil {
		t.Error("expected LeftSkill/RightSkill cleared")
	}

	if player.Stats.SkillPoints != 2 || player.Stats.StatsPoints != 3 || player.Stats.Vitality != 10 {
		t.Errorf("expected SkillPoints=2 StatsPoints=3 Vitality=10, got SkillPoints=%d StatsPoints=%d Vitality=%d",
			player.Stats.SkillPoints, player.Stats.StatsPoints, player.Stats.Vitality)
	}
}

func TestHandleItemCraftedPacketUpdatesWeaponAndGold(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Equipment: &d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: d2hero.ItemBatonApprenti},
		},
	})

	packet, err := d2netpacket.CreateItemCraftedPacket("p", d2hero.RecipeUpgradeBatonApprenti, d2hero.ItemBatonInitie, "Bâton de l'Initié", 50)
	if err != nil {
		t.Fatalf("test setup: CreateItemCraftedPacket failed: %v", err)
	}

	if err := client.handleItemCraftedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Equipment.RightHand.ItemCode != d2hero.ItemBatonInitie {
		t.Errorf("expected RightHand upgraded to %q, got %q", d2hero.ItemBatonInitie, player.Equipment.RightHand.ItemCode)
	}

	if player.Gold != 50 {
		t.Errorf("expected Gold=50, got %d", player.Gold)
	}
}

func TestHandleSkillLearnedPacketAddsSkillAndUpdatesPoints(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{SkillPoints: 2}})

	packet, err := d2netpacket.CreateSkillLearnedPacket("p", d2hero.SkillTraitDeFeu, 1)
	if err != nil {
		t.Fatalf("test setup: CreateSkillLearnedPacket failed: %v", err)
	}

	if err := client.handleSkillLearnedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if _, known := player.Skills[d2hero.SkillTraitDeFeu]; !known {
		t.Error("expected Trait de feu added to Skills")
	}

	if player.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints=1, got %d", player.Stats.SkillPoints)
	}
}

func TestHandleSkillEquippedPacketAssignsLeftAndRightIndependently(t *testing.T) {
	traitDeFeu := &d2hero.HeroSkill{}
	eclatDeGlace := &d2hero.HeroSkill{}
	client := clientWithPlayer("p", &d2mapentity.Player{
		Skills: map[int]*d2hero.HeroSkill{
			d2hero.SkillTraitDeFeu:   traitDeFeu,
			d2hero.SkillEclatDeGlace: eclatDeGlace,
		},
	})

	leftPacket, err := d2netpacket.CreateSkillEquippedPacket("p", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateSkillEquippedPacket failed: %v", err)
	}

	if err = client.handleSkillEquippedPacket(leftPacket); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rightPacket, err := d2netpacket.CreateSkillEquippedPacket("p", int(d2hero.SkillSlotRight), d2hero.SkillEclatDeGlace)
	if err != nil {
		t.Fatalf("test setup: CreateSkillEquippedPacket failed: %v", err)
	}

	if err := client.handleSkillEquippedPacket(rightPacket); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.LeftSkill != traitDeFeu {
		t.Error("expected LeftSkill assigned to Trait de feu")
	}

	if player.RightSkill != eclatDeGlace {
		t.Error("expected RightSkill assigned to Éclat de glace")
	}
}

func TestHandleSkillEquippedPacketUnknownSkillIsNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Skills: map[int]*d2hero.HeroSkill{}})

	packet, err := d2netpacket.CreateSkillEquippedPacket("p", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateSkillEquippedPacket failed: %v", err)
	}

	if err := client.handleSkillEquippedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.Players["p"].LeftSkill != nil {
		t.Error("expected LeftSkill untouched when the skill isn't known locally")
	}
}

func TestHandleSkillEquippedPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateSkillEquippedPacket("nobody", int(d2hero.SkillSlotLeft), d2hero.SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("test setup: CreateSkillEquippedPacket failed: %v", err)
	}

	// must not panic when the target isn't a known player.
	if err := client.handleSkillEquippedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedToStashPacketMovesTheItem(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Inventory: []string{d2hero.ItemPendentifArcane},
		Stash:     make([]string, 1),
	})

	packet, err := d2netpacket.CreateItemMovedToStashPacket("p", 0, 0, d2hero.ItemPendentifArcane)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToStashPacket failed: %v", err)
	}

	if err := client.handleItemMovedToStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Inventory[0] != "" {
		t.Errorf("expected the inventory slot cleared, got %q", player.Inventory[0])
	}

	if player.Stash[0] != d2hero.ItemPendentifArcane {
		t.Errorf("expected the item in the local stash copy, got %q", player.Stash[0])
	}
}

// TestHandleItemMovedToStashPacketOutOfRangeIsNoop is a regression test:
// Player.Inventory/Stash start empty (no full-sync packet exists on join,
// only these incremental move deltas), so a real client can genuinely
// receive an index beyond what it's ever seen -- this must not panic.
func TestHandleItemMovedToStashPacketOutOfRangeIsNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedToStashPacket("p", 0, 0, d2hero.ItemPendentifArcane)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToStashPacket failed: %v", err)
	}

	// must not panic indexing an empty Inventory/Stash.
	if err := client.handleItemMovedToStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedToStashPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedToStashPacket("nobody", 0, 0, d2hero.ItemPendentifArcane)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToStashPacket failed: %v", err)
	}

	if err := client.handleItemMovedToStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedFromStashPacketMovesTheItem(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Inventory: make([]string, 1),
		Stash:     []string{d2hero.ItemAnneauDuDebut},
	})

	packet, err := d2netpacket.CreateItemMovedFromStashPacket("p", 0, 0, d2hero.ItemAnneauDuDebut)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedFromStashPacket failed: %v", err)
	}

	if err := client.handleItemMovedFromStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Stash[0] != "" {
		t.Errorf("expected the stash slot cleared, got %q", player.Stash[0])
	}

	if player.Inventory[0] != d2hero.ItemAnneauDuDebut {
		t.Errorf("expected the item in the local inventory copy, got %q", player.Inventory[0])
	}
}

func TestHandleItemMovedFromStashPacketOutOfRangeIsNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedFromStashPacket("p", 0, 0, d2hero.ItemAnneauDuDebut)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedFromStashPacket failed: %v", err)
	}

	if err := client.handleItemMovedFromStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedFromStashPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedFromStashPacket("nobody", 0, 0, d2hero.ItemAnneauDuDebut)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedFromStashPacket failed: %v", err)
	}

	if err := client.handleItemMovedFromStashPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedToBeltPacketMovesThePotion(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Inventory: []string{d2hero.ItemPotionDeMana},
		Belt:      make([]string, 1),
	})

	packet, err := d2netpacket.CreateItemMovedToBeltPacket("p", 0, 0, d2hero.ItemPotionDeMana)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToBeltPacket failed: %v", err)
	}

	if err := client.handleItemMovedToBeltPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Inventory[0] != "" {
		t.Errorf("expected the inventory slot cleared, got %q", player.Inventory[0])
	}

	if player.Belt[0] != d2hero.ItemPotionDeMana {
		t.Errorf("expected the potion in the local belt copy, got %q", player.Belt[0])
	}
}

func TestHandleItemMovedToBeltPacketOutOfRangeIsNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedToBeltPacket("p", 0, 0, d2hero.ItemPotionDeMana)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToBeltPacket failed: %v", err)
	}

	// must not panic indexing an empty Inventory/Belt.
	if err := client.handleItemMovedToBeltPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedToBeltPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedToBeltPacket("nobody", 0, 0, d2hero.ItemPotionDeMana)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedToBeltPacket failed: %v", err)
	}

	if err := client.handleItemMovedToBeltPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleItemMovedFromBeltPacketMovesThePotion(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Inventory: make([]string, 1),
		Belt:      []string{d2hero.ItemPotionDeMana},
	})

	packet, err := d2netpacket.CreateItemMovedFromBeltPacket("p", 0, 0, d2hero.ItemPotionDeMana)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedFromBeltPacket failed: %v", err)
	}

	if err := client.handleItemMovedFromBeltPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Belt[0] != "" {
		t.Errorf("expected the belt slot cleared, got %q", player.Belt[0])
	}

	if player.Inventory[0] != d2hero.ItemPotionDeMana {
		t.Errorf("expected the potion in the local inventory copy, got %q", player.Inventory[0])
	}
}

func TestHandleItemMovedFromBeltPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateItemMovedFromBeltPacket("nobody", 0, 0, d2hero.ItemPotionDeMana)
	if err != nil {
		t.Fatalf("test setup: CreateItemMovedFromBeltPacket failed: %v", err)
	}

	if err := client.handleItemMovedFromBeltPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHandleSingleSkillRespecedPacketRemovesSkill(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Stats:  &d2hero.HeroStatsState{SkillPoints: 0},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillTraitDeFeu: {}, d2hero.SkillEclatDeGlace: {}},
	})

	packet, err := d2netpacket.CreateSingleSkillRespecedPacket("p", d2hero.SkillTraitDeFeu, 1)
	if err != nil {
		t.Fatalf("test setup: CreateSingleSkillRespecedPacket failed: %v", err)
	}

	if err := client.handleSingleSkillRespecedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if _, known := player.Skills[d2hero.SkillTraitDeFeu]; known {
		t.Error("expected Trait de feu forgotten")
	}

	if _, known := player.Skills[d2hero.SkillEclatDeGlace]; !known {
		t.Error("expected Éclat de glace to remain learned")
	}

	if player.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints=1, got %d", player.Stats.SkillPoints)
	}
}

func TestHandlePotionUsedPacketUpdatesMana(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{Mana: 0}})

	packet, err := d2netpacket.CreatePotionUsedPacket("p", 15)
	if err != nil {
		t.Fatalf("test setup: CreatePotionUsedPacket failed: %v", err)
	}

	if err := client.handlePotionUsedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.Players["p"].Stats.Mana != 15 {
		t.Errorf("expected Mana=15, got %d", client.Players["p"].Stats.Mana)
	}
}

// TestHandlePlayerDamagedPacketUpdatesHealth also covers Mana: a regression
// test for Bouclier de mana draining Mana instead of Health on the same
// hit this packet reports, which the client previously never learned about
// since only HP was ever carried.
func TestHandlePlayerDamagedPacketUpdatesHealth(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{Health: 50, Mana: 20}})

	packet, err := d2netpacket.CreatePlayerDamagedPacket("p", 30, 5, false)
	if err != nil {
		t.Fatalf("test setup: CreatePlayerDamagedPacket failed: %v", err)
	}

	if err := client.handlePlayerDamagedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.Players["p"].Stats.Health != 30 {
		t.Errorf("expected Health=30, got %d", client.Players["p"].Stats.Health)
	}

	if client.Players["p"].Stats.Mana != 5 {
		t.Errorf("expected Mana=5, got %d", client.Players["p"].Stats.Mana)
	}
}

func TestHandleGoldAwardedPacketUpdatesGold(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Gold: 0})

	packet, err := d2netpacket.CreateGoldAwardedPacket("p", 15)
	if err != nil {
		t.Fatalf("test setup: CreateGoldAwardedPacket failed: %v", err)
	}

	if err := client.handleGoldAwardedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if client.Players["p"].Gold != 15 {
		t.Errorf("expected Gold=15, got %d", client.Players["p"].Gold)
	}
}

func TestHandleExperienceAwardedPacketUpdatesLevelState(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{Stats: &d2hero.HeroStatsState{}})

	packet, err := d2netpacket.CreateExperienceAwardedPacket("p", 350, 3, 2, 10)
	if err != nil {
		t.Fatalf("test setup: CreateExperienceAwardedPacket failed: %v", err)
	}

	if err := client.handleExperienceAwardedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stats := client.Players["p"].Stats
	if stats.Experience != 350 || stats.Level != 3 || stats.SkillPoints != 2 || stats.StatsPoints != 10 {
		t.Errorf("expected Experience=350 Level=3 SkillPoints=2 StatsPoints=10, got Experience=%d Level=%d SkillPoints=%d StatsPoints=%d",
			stats.Experience, stats.Level, stats.SkillPoints, stats.StatsPoints)
	}
}

func TestHandlePlayerTeleportedPacketMovesPlayer(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreatePlayerTeleportedPacket("p", 10, 20)
	if err != nil {
		t.Fatalf("test setup: CreatePlayerTeleportedPacket failed: %v", err)
	}

	if err := client.handlePlayerTeleportedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	pos := client.Players["p"].GetPosition()
	if pos.X() != 10 || pos.Y() != 20 {
		t.Errorf("expected position (10, 20), got (%v, %v)", pos.X(), pos.Y())
	}
}

func TestHandleSkillPointInvestedPacketUpdatesInvestedPoints(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{
		Stats:  &d2hero.HeroStatsState{SkillPoints: 1},
		Skills: map[int]*d2hero.HeroSkill{d2hero.SkillMaitriseElementaire: {SkillPoints: 1}},
	})

	packet, err := d2netpacket.CreateSkillPointInvestedPacket("p", d2hero.SkillMaitriseElementaire, 2, 0)
	if err != nil {
		t.Fatalf("test setup: CreateSkillPointInvestedPacket failed: %v", err)
	}

	if err := client.handleSkillPointInvestedPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	player := client.Players["p"]
	if player.Skills[d2hero.SkillMaitriseElementaire].SkillPoints != 2 {
		t.Errorf("expected 2 points invested, got %d", player.Skills[d2hero.SkillMaitriseElementaire].SkillPoints)
	}

	if player.Stats.SkillPoints != 0 {
		t.Errorf("expected SkillPoints drained to 0, got %d", player.Stats.SkillPoints)
	}
}

// TestHandlePlayerDisconnectionPacketUnknownPlayerNoop is a regression test:
// a disconnection notification for a player ID not in g.Players (e.g. a
// duplicate/replayed packet delivered after the first one already removed
// and deleted that entry) used to look up g.Players without a `found`
// check, then pass the resulting nil *d2mapentity.Player straight into
// MapEngine.RemoveEntity -- a classic Go typed-nil-in-interface gotcha:
// RemoveEntity's own `entity == nil` guard doesn't catch a nil concrete
// pointer wrapped in a non-nil interface value, so it panicked dereferencing
// the nil pointer at entity.ID(). g.MapEngine is deliberately left nil here:
// the fixed code must never touch it for an unknown player ID.
func TestHandlePlayerDisconnectionPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreatePlayerDisconnectRequestPacket("nobody")
	if err != nil {
		t.Fatalf("test setup: CreatePlayerDisconnectRequestPacket failed: %v", err)
	}

	if err := client.handlePlayerDisconnectionPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, found := client.Players["p"]; !found {
		t.Error("expected the unrelated known player to be left untouched")
	}
}

// TestHandleMovePlayerPacketUnknownPlayerNoop is a regression test for the
// same class of bug: a MovePlayer packet for an unknown player ID used to
// dereference a nil *d2mapentity.Player at player.SetPath.
func TestHandleMovePlayerPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateMovePlayerPacket("nobody", 0, 0, 1, 1)
	if err != nil {
		t.Fatalf("test setup: CreateMovePlayerPacket failed: %v", err)
	}

	if err := client.handleMovePlayerPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestHandleCastSkillPacketUnknownPlayerNoop is a regression test for the
// same class of bug: a CastSkill packet for an unknown source entity ID
// used to dereference a nil *d2mapentity.Player at player.StopMoving.
func TestHandleCastSkillPacketUnknownPlayerNoop(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	packet, err := d2netpacket.CreateCastPacket("nobody", d2hero.SkillTraitDeFeu, 0, 0)
	if err != nil {
		t.Fatalf("test setup: CreateCastPacket failed: %v", err)
	}

	if err := client.handleCastSkillPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestHandleCastSkillPacketDevilSkillDoesNotPanic is a regression test for
// a real, previously-uncaught crash: Skill.Details is the stock skills.txt
// table, which has no entry for any Devil skill (SkillTraitDeFeu etc.) --
// skillRecord used to come back nil and get dereferenced immediately
// (skillRecord.Cltmissile inside createMissileEntities), guaranteed to
// panic the instant a real, *known* player cast any Devil skill. The
// existing TestHandleCastSkillPacketUnknownPlayerNoop above also uses a
// Devil skill ID, but targets an unknown player, so it returns early
// before ever reaching the vulnerable code -- this test uses a real,
// known player instead, to actually exercise that path.
func TestHandleCastSkillPacketDevilSkillDoesNotPanic(t *testing.T) {
	client := clientWithPlayer("p", &d2mapentity.Player{})

	// g.asset.Records.Skill.Details[id] needs a real, non-nil AssetManager
	// to reach at all (a nil *AssetManager panics on the first field
	// access, before the map lookup this test actually cares about) -- a
	// freshly constructed one has an empty Skill.Details map, and a Go map
	// read on a missing key returns the zero value rather than panicking,
	// which is exactly the real-world case being tested here (Devil skill
	// IDs are never in this stock table).
	assetManager, err := d2asset.NewAssetManager(d2util.LogLevelNone)
	if err != nil {
		t.Fatalf("test setup: NewAssetManager failed: %v", err)
	}

	client.asset = assetManager

	packet, err := d2netpacket.CreateCastPacket("p", d2hero.SkillTraitDeFeu, 5, 5)
	if err != nil {
		t.Fatalf("test setup: CreateCastPacket failed: %v", err)
	}

	if err := client.handleCastSkillPacket(packet); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
