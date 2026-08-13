package d2client

import (
	"testing"
	"time"

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
