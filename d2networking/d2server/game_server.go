package d2server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapengine"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapgen"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2clientconnectiontype"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2server/d2tcpclientconnection"
	"github.com/OpenDiablo2/OpenDiablo2/d2script"
)

const logPrefix = "Game Server"

const (
	port                   = "6669"
	chunkSize          int = 4096 // nolint:deadcode,unused,varcheck // WIP
	subtilesPerTile        = 5
	middleOfTileOffset     = 3
)

// ponytail: flat hit radius instead of real per-skill reach, and no attack
// rating/defense to-hit roll -- a hit within range always lands. Enough to
// validate the hit-resolution pipeline end to end; see ROADMAP.md Phase 1.
const meleeHitRadiusSubtiles = 3

// baseSortDamage is the fallback base_sort damage used when the cast
// skill isn't in d2hero.DevilSkills (or no connection/stats resolved).
const baseSortDamage = 4

// baseCastCooldown is the time between casts at 0 Dexterity.
const baseCastCooldown = 800 * time.Millisecond

// minCastCooldown is the floor cooldown can never go below, however high
// Dexterity gets.
const minCastCooldown = 200 * time.Millisecond

// castCooldownFor returns the time a caster with the given Dexterity must
// wait between casts, per "Dexterity ... réduit le temps de récupération
// après un impact" (devil_game_design_reference.md §6).
//
// ponytail: continuous scaling instead of the design's discrete
// "breakpoints d'incantation" -- those are frame-count thresholds tied to
// real cast animation data, which doesn't exist yet (ROADMAP.md Phase 2).
// This is a placeholder with the same shape (higher Dexterity -> faster
// casts, floored so it never reaches zero), not the real formula.
func castCooldownFor(dexterity int) time.Duration {
	cooldown := baseCastCooldown / time.Duration(1+dexterity/25)
	if cooldown < minCastCooldown {
		return minCastCooldown
	}

	return cooldown
}

// defaultManaCost is the fallback mana cost used when the cast skill isn't
// in d2hero.DevilSkills.
const defaultManaCost = 2

// baseManaRegenPerSecond is how much mana regenerates per second at 0
// Energy; manaRegenPerSecond scales it up from there.
const baseManaRegenPerSecond = 1.0

// manaRegenPerSecond returns how much mana regenerates per second for the
// given Energy, per "Régénération de mana ... boostée par l'Energy"
// (devil_game_design_reference.md §6).
//
// ponytail: the design doc states the relationship (higher Energy -> faster
// regen) but not a formula -- this is an untuned placeholder, not a real
// balance value.
func manaRegenPerSecond(energy int) float64 {
	return baseManaRegenPerSecond + float64(energy)/50
}

// canCastNow reports whether sourceEntityID may deal damage with a cast of
// skillID right now: they must be off cooldown (castCooldownFor) and, if
// they're a resolved player, have enough mana (d2hero.SkillManaCost) after applying
// regen (manaRegenPerSecond) since their last cast. Mana is deducted and
// this moment recorded as their last cast if both checks pass.
//
// A caster still on cooldown or short on mana gets no hit resolution at
// all -- same as if they'd cast at nothing (see resolveMeleeHit); the
// cast's visual effect still plays for everyone since that's relayed
// independently of this gate.
func (g *GameServer) canCastNow(sourceEntityID string, skillID int) bool {
	now := g.clock()

	g.Lock()
	defer g.Unlock()

	last, hasCastBefore := g.lastCastAt[sourceEntityID]

	if hasCastBefore && now.Sub(last) < castCooldownFor(g.dexterityOf(sourceEntityID)) {
		return false
	}

	if state := g.playerStateOf(sourceEntityID); state != nil && state.Stats != nil {
		if hasCastBefore {
			regenPerSecond := manaRegenPerSecond(effectiveEnergy(state))

			if accel, learned := state.Skills[d2hero.SkillRegenerationAcceleree]; learned {
				if percent := d2hero.RegenerationAccellereePercent(accel.SkillPoints); percent > 0 {
					regenPerSecond += regenPerSecond * float64(percent) / 100
				}
			}

			regen := int(regenPerSecond * now.Sub(last).Seconds())
			state.Stats.Mana = min(state.Stats.Mana+regen, state.Stats.MaxMana)
		}

		cost := d2hero.SkillManaCost(skillID, defaultManaCost)
		if state.Stats.Mana < cost {
			return false
		}

		state.Stats.Mana -= cost
	}

	g.lastCastAt[sourceEntityID] = now

	return true
}

// playerStateOf returns sourceEntityID's HeroState, or nil if they aren't a
// connected player.
func (g *GameServer) playerStateOf(sourceEntityID string) *d2hero.HeroState {
	connection, ok := g.connections[sourceEntityID]
	if !ok {
		return nil
	}

	return connection.GetPlayerState()
}

// dexterityOf returns sourceEntityID's effective Dexterity (see
// effectiveDexterity), or 0 if it isn't a connected player or has no stats
// resolved.
func (g *GameServer) dexterityOf(sourceEntityID string) int {
	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return 0
	}

	return effectiveDexterity(state)
}

// equippedItemCodes returns the item code of every equipment slot that
// could carry a Devil item bonus (RightHand/Amulet/Ring). Unequipped slots
// contribute "" harmlessly -- d2hero's Item*Bonus lookups return 0/fallback
// for unknown codes.
func equippedItemCodes(state *d2hero.HeroState) []string {
	return []string{
		state.Equipment.RightHand.GetItemCode(),
		state.Equipment.Amulet.GetItemCode(),
		state.Equipment.Ring.GetItemCode(),
	}
}

// effectiveEnergy returns state's own Energy plus every equipped item's
// Energy-specific bonus (d2hero.ItemEnergyBonus) and all-attributes bonus
// (d2hero.ItemAllAttributesBonus) -- e.g. Bâton de l'Apprenti's "+5
// Energy", Pendentif Arcane's "+3 Energy", and Anneau du Début's "+2 à
// tous les attributs" (devil_mage_character_design.md §5), none of which
// previously had any effect on gameplay: resolveAttackDamage and
// manaRegenPerSecond both read state.Stats.Energy directly, ignoring
// equipment entirely.
func effectiveEnergy(state *d2hero.HeroState) int {
	total := state.Stats.Energy

	for _, code := range equippedItemCodes(state) {
		total += d2hero.ItemEnergyBonus(code) + d2hero.ItemAllAttributesBonus(code)
	}

	return total
}

// effectiveDexterity returns state's own Dexterity plus every equipped
// item's all-attributes bonus (d2hero.ItemAllAttributesBonus) -- no Devil
// item grants a Dexterity-only bonus yet, only the all-attributes shape
// (Anneau du Début).
func effectiveDexterity(state *d2hero.HeroState) int {
	total := state.Stats.Dexterity

	for _, code := range equippedItemCodes(state) {
		total += d2hero.ItemAllAttributesBonus(code)
	}

	return total
}

// aiTickInterval is how often the monster AI loop reevaluates.
const aiTickInterval = 200 * time.Millisecond

// monsterAggroRadiusSubtiles/monsterAttackRangeSubtiles/monsterAttackCooldown/
// monsterAttackDamage are the fallback values used for any monster whose
// key isn't in d2hero.MonsterArchetypes -- true for every monster today, no
// Devil monster data exists yet (ROADMAP.md Phase 4). advanceMonsterAI/
// tryMonsterAttack already look up d2hero.MonsterArchetypes first, so
// registering a real archetype there overrides these per monster type
// without any further server-side changes.
const (
	monsterAggroRadiusSubtiles = 8
	monsterAttackRangeSubtiles = 2
	monsterAttackCooldown      = 1200 * time.Millisecond
	monsterAttackDamage        = 3
)

// armureDeGlaceDamageReductionPercent/armureDeGlaceSlowDuration: while
// Armure de glace is active (HeroStatsState.ArmureDeGlaceActive), incoming
// monster attacks are reduced by this percent and the attacking NPC is
// slowed for this long (devil_game_design_reference.md §7 Ésotérisme:
// "Réduit les dégâts reçus et ralentit les attaquants au contact").
//
// ponytail: flat placeholder numbers -- the design doesn't specify exact
// values.
const (
	armureDeGlaceDamageReductionPercent = 30
	armureDeGlaceSlowDuration           = 3 * time.Second
)

// eclatDeGlaceSlowDuration is how long a hit NPC's movement speed stays
// reduced (d2mapentity.NPC.ApplySlow) after being hit by Éclat de glace.
//
// ponytail: a flat duration -- the design doesn't specify one, and there's
// no per-skill duration data model yet (only base_sort/mana cost exist on
// DevilSkillDef today).
const eclatDeGlaceSlowDuration = 3 * time.Second

// skillEclairEnChaineTargets/eclairEnChaineChainRadius: Éclair en chaîne
// hits its first target, then jumps to up to this many more nearby
// killable NPCs, each within eclairEnChaineChainRadius subtiles of the
// previous one (devil_game_design_reference.md §7: "Foudre qui rebondit
// sur 3 cibles").
const (
	skillEclairEnChaineTargets = 3
	eclairEnChaineChainRadius  = 6
)

// novaDeGivreRadiusSubtiles/tempeteDeLamesRadiusSubtiles: how many subtiles
// around the caster these self-centered AoE spells hit
// (devil_game_design_reference.md §7: "Explosion de froid en zone autour du
// Mage" / "Invoque des lames de mana orbitant autour du Mage" -- the
// "orbiting/periodic" part of Tempête de lames isn't modeled, same
// simplification as Tempête statique's missing "persistante" zone) --
// centered on the caster, not the cast's targeted position.
const (
	novaDeGivreRadiusSubtiles    = 5
	tempeteDeLamesRadiusSubtiles = 4
)

// selfCenteredAoeRadiusSubtiles maps each self-centered-AoE skill to its
// radius, so resolveMeleeHit can dispatch to resolveSelfCenteredAoeHit
// without a growing if-chain -- mirrors aoeAtTargetRadiusSubtiles below for
// the target-centered equivalent.
var selfCenteredAoeRadiusSubtiles = map[int]float64{
	d2hero.SkillNovaDeGivre:    novaDeGivreRadiusSubtiles,
	d2hero.SkillTempeteDeLames: tempeteDeLamesRadiusSubtiles,
}

// bouleDeFeuRadiusSubtiles/tempeteStatiqueRadiusSubtiles/orbeGlacialeRadiusSubtiles/meteoreRadiusSubtiles:
// how many subtiles around the cast's targeted position these
// AoE-at-target spells hit ("Projectile AoE, dégâts feu élevés" / "Invoque
// une zone d'éclair persistante" -- see SkillTempeteStatique's doc comment
// for why the "persistante" part isn't modeled yet -- / "explose en large
// AoE de froid" / "Frappe retardée sur une zone, dégâts feu massifs" -- see
// SkillMeteore's doc comment for why the delay isn't modeled either).
const (
	bouleDeFeuRadiusSubtiles      = 3
	tempeteStatiqueRadiusSubtiles = 4
	orbeGlacialeRadiusSubtiles    = 6 // "large AoE" -- the biggest radius of the three
	meteoreRadiusSubtiles         = 4
)

// aoeAtTargetRadiusSubtiles maps each AoE-at-target-position Élémentalisme
// skill to its radius, so resolveMeleeHit can dispatch to resolveAoeHit
// without a growing if-chain as more such spells are added (Apocalypse is
// this shape too -- ROADMAP.md Phase 2).
var aoeAtTargetRadiusSubtiles = map[int]float64{
	d2hero.SkillBouleDeFeu:      bouleDeFeuRadiusSubtiles,
	d2hero.SkillTempeteStatique: tempeteStatiqueRadiusSubtiles,
	d2hero.SkillOrbeGlaciale:    orbeGlacialeRadiusSubtiles,
	d2hero.SkillMeteore:         meteoreRadiusSubtiles,
}

// telekinesieRadiusSubtiles/telekinesieKnockbackDistance: Télékinésie finds
// every killable NPC within this many subtiles of the cast's targeted
// position, and pushes each away from the caster by this many subtiles
// (devil_game_design_reference.md §7 Arcane: "Repousse les entités").
const (
	telekinesieRadiusSubtiles    = 3
	telekinesieKnockbackDistance = 4
)

// runMonsterAILoop periodically advances monster AI. Meant to be started as
// a goroutine; returns once the server is stopped.
func (g *GameServer) runMonsterAILoop() {
	ticker := time.NewTicker(aiTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.advanceMonsterAI()
		}
	}
}

// advanceMonsterAI gives every killable, living NPC on the map a chance to
// notice and react to the nearest connected player: chase if aggroed but
// out of attack range, or attack if in range (subject to its own cooldown,
// see tryMonsterAttack). NPCs with no player within monsterAggroRadiusSubtiles
// are left alone entirely.
func (g *GameServer) advanceMonsterAI() {
	if len(g.mapEngines) == 0 {
		return
	}

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		npcPos := npc.GetPosition()

		playerID, playerPos, found := g.nearestPlayer(npcPos)
		if !found {
			continue
		}

		dist := npcPos.Distance(&playerPos.Vector)
		monsterKey := npc.MonsterKey()

		if dist > d2hero.MonsterAggroRadiusSubtiles(monsterKey, monsterAggroRadiusSubtiles) {
			continue
		}

		if dist > d2hero.MonsterAttackRangeSubtiles(monsterKey, monsterAttackRangeSubtiles) {
			npc.ChasePlayer(playerPos, g.clock())
			continue
		}

		g.tryMonsterAttack(npc.ID(), playerID)
	}
}

// nearestPlayer returns the connected player closest to from, reading live
// position off their PlayerState (the server has no map-entity for
// players -- see GameServer.connections).
func (g *GameServer) nearestPlayer(from d2vector.Position) (playerID string, pos d2vector.Position, found bool) {
	nearestDist := math.MaxFloat64

	for id, connection := range g.connections {
		state := connection.GetPlayerState()
		if state == nil {
			continue
		}

		candidate := d2vector.NewPosition(state.X, state.Y)
		if dist := from.Distance(&candidate.Vector); dist < nearestDist {
			playerID, pos, nearestDist, found = id, candidate, dist, true
		}
	}

	return playerID, pos, found
}

// tryMonsterAttack applies monsterAttackDamage to playerID's HP and
// broadcasts the result, unless npcID is still on its own attack cooldown.
func (g *GameServer) tryMonsterAttack(npcID, playerID string) {
	now := g.clock()

	var monsterKey string
	if npc := g.npcByID(npcID); npc != nil {
		monsterKey = npc.MonsterKey()
	}

	cooldown := d2hero.MonsterAttackCooldown(monsterKey, monsterAttackCooldown)

	g.Lock()
	last, attacked := g.lastMonsterAttackAt[npcID]

	if attacked && now.Sub(last) < cooldown {
		g.Unlock()
		return
	}

	g.lastMonsterAttackAt[npcID] = now
	g.Unlock()

	connection, ok := g.connections[playerID]
	if !ok {
		return
	}

	state := connection.GetPlayerState()
	if state == nil || state.Stats == nil {
		return
	}

	// ponytail: every monster attack treated as Fire damage -- no per-monster
	// element data exists yet (ROADMAP.md Phase 4). Enough to prove
	// resistances actually mitigate something.
	damage := d2hero.MitigateDamage(
		d2hero.MonsterAttackDamage(monsterKey, monsterAttackDamage), d2hero.CapResistance(state.Stats.FireResist))

	switch {
	case state.Stats.IsMagicImmune(now):
		// Éveil du Nexus: full immunity overrides everything else below,
		// including Armure de glace's partial reduction.
		damage = 0
	case state.Stats.ArmureDeGlaceActive:
		damage = damage * (100 - armureDeGlaceDamageReductionPercent) / 100

		if attacker := g.npcByID(npcID); attacker != nil {
			until := now.Add(armureDeGlaceSlowDuration)
			attacker.ApplySlow(until)
			g.broadcastNPCStatusEffect(attacker, d2netpacket.NPCStatusSlowed, until)
		}
	}

	died := state.Stats.ApplyDamageWithManaShield(damage)

	if died && g.tryTranscend(playerID, state) {
		died = false
	}

	packet, err := d2netpacket.CreatePlayerDamagedPacket(playerID, state.Stats.Health, died)
	if err != nil {
		g.Errorf("CreatePlayerDamagedPacket: %v", err)
		return
	}

	g.sendPacketToClients(packet)
}

// transcendanceHealPercent is how much of MaxHealth Transcendance restores
// when it saves the Mage from death ("le Mage se régénère").
//
// ponytail: the design names the effect but not an amount; a full heal
// (100%) is a placeholder pending real balance numbers -- reasonable for a
// rare death save.
const transcendanceHealPercent = 100

// transcendanceCooldown is how long Transcendance's death-save must wait
// before it can trigger again.
//
// ponytail: "une fois par zone" (devil_game_design_reference.md §7
// Ésotérisme) has no real meaning yet -- there's no zone-transition event
// to reset it against. A long flat cooldown stands in until zone
// transitions exist.
const transcendanceCooldown = 5 * time.Minute

// tryTranscend checks whether playerID has learned Transcendance and isn't
// on cooldown; if so, it heals them (transcendanceHealPercent of MaxHealth)
// and reports true so the caller can override a death. A no-op (false)
// otherwise.
func (g *GameServer) tryTranscend(playerID string, state *d2hero.HeroState) bool {
	if _, learned := state.Skills[d2hero.SkillTranscendance]; !learned {
		return false
	}

	now := g.clock()

	g.Lock()
	last, used := g.lastTranscendanceAt[playerID]

	if used && now.Sub(last) < transcendanceCooldown {
		g.Unlock()
		return false
	}

	g.lastTranscendanceAt[playerID] = now
	g.Unlock()

	state.Stats.Heal(state.Stats.MaxHealth * transcendanceHealPercent / 100)

	return true
}

// awardGold rolls a gold drop (d2hero.RollGoldDrop) and credits it to
// playerID, broadcasting their new total. A no-op if playerID isn't a
// resolved connected player.
func (g *GameServer) awardGold(playerID string) {
	state := g.playerStateOf(playerID)
	if state == nil {
		return
	}

	state.Gold += d2hero.RollGoldDrop()

	packet, err := d2netpacket.CreateGoldAwardedPacket(playerID, state.Gold)
	if err != nil {
		g.Errorf("CreateGoldAwardedPacket: %v", err)
		return
	}

	g.sendPacketToClients(packet)
}

// awardExperience rolls an experience drop (d2hero.RollExperienceDrop),
// grants it to playerID (possibly leveling them up -- see
// d2hero.HeroStatsState.GrantExperience), and broadcasts their new
// experience/level state. A no-op if playerID isn't a resolved connected
// player or has no stats.
func (g *GameServer) awardExperience(playerID string) {
	state := g.playerStateOf(playerID)
	if state == nil || state.Stats == nil {
		return
	}

	state.Stats.GrantExperience(d2hero.RollExperienceDrop())

	packet, err := d2netpacket.CreateExperienceAwardedPacket(
		playerID, state.Stats.Experience, state.Stats.Level, state.Stats.SkillPoints, state.Stats.StatsPoints)
	if err != nil {
		g.Errorf("CreateExperienceAwardedPacket: %v", err)
		return
	}

	g.sendPacketToClients(packet)
}

// absorptionEnergieManaRestorePercent is how much of MaxMana Absorption
// d'énergie restores per kill, if learned (devil_game_design_reference.md
// §7 Ésotérisme: "Chaque entité tuée restaure un % de mana").
//
// ponytail: the design names the effect but not a percentage; 20% is a
// placeholder pending real balance numbers.
const absorptionEnergieManaRestorePercent = 20

// restoreManaOnKill restores absorptionEnergieManaRestorePercent of
// playerID's own MaxMana if they've learned Absorption d'énergie (a true
// passive -- there's no cast/dispatch for it, this just runs on every kill).
// A no-op otherwise, or if playerID isn't a resolved connected player.
func (g *GameServer) restoreManaOnKill(playerID string) {
	state := g.playerStateOf(playerID)
	if state == nil || state.Stats == nil {
		return
	}

	if _, learned := state.Skills[d2hero.SkillAbsorptionEnergie]; !learned {
		return
	}

	state.Stats.RestoreMana(state.Stats.MaxMana * absorptionEnergieManaRestorePercent / 100)

	// Correction (août 2026): the restore itself was never broadcast either
	// -- reuses PotionUsedPacket (it just carries the player's new Mana
	// total, nothing potion-specific) rather than adding a redundant packet
	// type for the exact same shape.
	restoredPacket, err := d2netpacket.CreatePotionUsedPacket(playerID, state.Stats.Mana)
	if err != nil {
		g.Errorf("CreatePotionUsedPacket: %v", err)
		return
	}

	g.sendPacketToClients(restoredPacket)
}

// resolveUsePotion unmarshals a UsePotionRequestPacket and, if
// HeroState.UsePotion succeeds (valid slot, not empty, hero has stats),
// broadcasts the caster's new Mana total via PotionUsedPacket. Silently
// does nothing on any failure (unknown player, bad slot, empty slot) --
// same "no hit resolution at all" shape as a cast that can't resolve, see
// resolveMeleeHit.
func (g *GameServer) resolveUsePotion(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalUsePotionRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveUsePotion: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil {
		return
	}

	if err := state.UsePotion(requestPacket.BeltIndex); err != nil {
		return
	}

	usedPacket, err := d2netpacket.CreatePotionUsedPacket(requestPacket.SourceEntityID, state.Stats.Mana)
	if err != nil {
		g.Errorf("CreatePotionUsedPacket: %v", err)
		return
	}

	g.sendPacketToClients(usedPacket)
}

// resolveLearnSkill unmarshals a LearnSkillRequestPacket and, if
// HeroState.LearnSkill succeeds (level gate, has a point to spend, not
// already known), broadcasts the result via SkillLearnedPacket. Silently
// does nothing on any failure -- same "no hit resolution at all" shape as
// a cast/potion that can't resolve, see resolveMeleeHit/resolveUsePotion.
func (g *GameServer) resolveLearnSkill(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalLearnSkillRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveLearnSkill: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil {
		return
	}

	if err := state.LearnSkill(requestPacket.SkillID); err != nil {
		return
	}

	learnedPacket, err := d2netpacket.CreateSkillLearnedPacket(
		requestPacket.SourceEntityID, requestPacket.SkillID, state.Stats.SkillPoints)
	if err != nil {
		g.Errorf("CreateSkillLearnedPacket: %v", err)
		return
	}

	g.sendPacketToClients(learnedPacket)
}

// resolveRespecSkills unmarshals a RespecSkillsRequestPacket and forgets
// every skill the caster has learned (HeroState.RespecSkills -- "Respec
// partiel"), broadcasting their refunded SkillPoints. Unlike
// resolveLearnSkill/resolveUsePotion, RespecSkills has no failure mode of
// its own (forgetting zero skills is a valid no-op) -- the only way this
// silently does nothing is an unresolved caster.
func (g *GameServer) resolveRespecSkills(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalRespecSkillsRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveRespecSkills: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	state.RespecSkills()

	respecedPacket, err := d2netpacket.CreateSkillsRespecedPacket(requestPacket.SourceEntityID, state.Stats.SkillPoints)
	if err != nil {
		g.Errorf("CreateSkillsRespecedPacket: %v", err)
		return
	}

	g.sendPacketToClients(respecedPacket)
}

// resolveRespecSingleSkill unmarshals a RespecSingleSkillRequestPacket and
// forgets exactly the requested skill (HeroState.RespecSingleSkill --
// "Glyphe d'oubli"), broadcasting the forgotten skill and the caster's
// refunded SkillPoints. Silently does nothing if the skill wasn't learned
// or the caster isn't resolved.
func (g *GameServer) resolveRespecSingleSkill(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalRespecSingleSkillRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveRespecSingleSkill: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	if _, err := state.RespecSingleSkill(requestPacket.SkillID); err != nil {
		return
	}

	respecedPacket, err := d2netpacket.CreateSingleSkillRespecedPacket(
		requestPacket.SourceEntityID, requestPacket.SkillID, state.Stats.SkillPoints)
	if err != nil {
		g.Errorf("CreateSingleSkillRespecedPacket: %v", err)
		return
	}

	g.sendPacketToClients(respecedPacket)
}

// resolveInvestSkillPoint unmarshals an InvestSkillPointRequestPacket and,
// if the caster already knows the skill and has a point to spend, adds
// another point to it (HeroState.InvestSkillPoint -- e.g. Maîtrise
// élémentaire's "+% dégâts élémentaires par point"), broadcasting the
// skill's new invested total and the caster's remaining SkillPoints.
func (g *GameServer) resolveInvestSkillPoint(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalInvestSkillPointRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveInvestSkillPoint: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	if err := state.InvestSkillPoint(requestPacket.SkillID); err != nil {
		return
	}

	investedPacket, err := d2netpacket.CreateSkillPointInvestedPacket(
		requestPacket.SourceEntityID, requestPacket.SkillID,
		state.Skills[requestPacket.SkillID].SkillPoints, state.Stats.SkillPoints)
	if err != nil {
		g.Errorf("CreateSkillPointInvestedPacket: %v", err)
		return
	}

	g.sendPacketToClients(investedPacket)
}

// attributeValue returns attr's current value on stats (Strength/Energy/
// Dexterity/Vitality), or 0 for an unrecognized attribute. Shared by
// resolveSpendAttributePoint/resolveRespecSingleAttributePoint so both can
// report the attribute's new value without duplicating the switch.
func attributeValue(stats *d2hero.HeroStatsState, attr d2hero.Attribute) int {
	switch attr {
	case d2hero.AttributeStrength:
		return stats.Strength
	case d2hero.AttributeEnergy:
		return stats.Energy
	case d2hero.AttributeDexterity:
		return stats.Dexterity
	case d2hero.AttributeVitality:
		return stats.Vitality
	default:
		return 0
	}
}

// resolveSpendAttributePoint unmarshals a SpendAttributePointRequestPacket
// and, if the caster has a point available, spends it on the requested
// attribute (HeroStatsState.SpendAttributePoint), broadcasting the
// attribute's new value and the caster's remaining StatsPoints.
func (g *GameServer) resolveSpendAttributePoint(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalSpendAttributePointRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveSpendAttributePoint: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	attr := d2hero.Attribute(requestPacket.Attribute)
	if err := state.Stats.SpendAttributePoint(attr); err != nil {
		return
	}

	spentPacket, err := d2netpacket.CreateAttributePointSpentPacket(
		requestPacket.SourceEntityID, requestPacket.Attribute, attributeValue(state.Stats, attr), state.Stats.StatsPoints)
	if err != nil {
		g.Errorf("CreateAttributePointSpentPacket: %v", err)
		return
	}

	g.sendPacketToClients(spentPacket)
}

// resolveRespecSingleAttributePoint unmarshals a
// RespecSingleAttributePointRequestPacket and, if the caster has a point
// spent on the requested attribute, refunds it
// (HeroState.RespecSingleAttributePoint -- the attribute half of "Glyphe
// d'oubli"), broadcasting the attribute's new value and the caster's new
// StatsPoints total.
func (g *GameServer) resolveRespecSingleAttributePoint(packet d2netpacket.NetPacket) {
	requestPacket, err := d2netpacket.UnmarshalRespecSingleAttributePointRequest(packet.PacketData)
	if err != nil {
		g.Errorf("resolveRespecSingleAttributePoint: %v", err)
		return
	}

	state := g.playerStateOf(requestPacket.SourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	attr := d2hero.Attribute(requestPacket.Attribute)
	if err := state.RespecSingleAttributePoint(attr); err != nil {
		return
	}

	respecedPacket, err := d2netpacket.CreateSingleAttributePointRespecedPacket(
		requestPacket.SourceEntityID, requestPacket.Attribute, attributeValue(state.Stats, attr), state.Stats.StatsPoints)
	if err != nil {
		g.Errorf("CreateSingleAttributePointRespecedPacket: %v", err)
		return
	}

	g.sendPacketToClients(respecedPacket)
}

// resolveMeleeHit checks for a killable NPC near the cast's target position
// and, if one is found within range, applies damage and broadcasts the
// result. Targeting is purely proximity-based for now -- see the constants
// above and ROADMAP.md Phase 1.
func (g *GameServer) resolveMeleeHit(packet d2netpacket.NetPacket) {
	if len(g.mapEngines) == 0 {
		return
	}

	castPacket, err := d2netpacket.UnmarshalCast(packet.PacketData)
	if err != nil {
		g.Errorf("resolveMeleeHit: %v", err)
		return
	}

	if !g.canCastNow(castPacket.SourceEntityID, castPacket.SkillID) {
		return
	}

	// Résonance magique: "chaque sort lancé augmente les dégâts du suivant"
	// -- every dispatched cast re-arms the bonus, regardless of which skill
	// it is or whether the caster even knows Résonance magique (consuming
	// it in resolveAttackDamage is what actually gates that).
	if state := g.playerStateOf(castPacket.SourceEntityID); state != nil && state.Stats != nil {
		state.Stats.ApplyResonanceMagiqueBonus(g.clock().Add(resonanceMagiqueWindow))
	}

	if radius, ok := selfCenteredAoeRadiusSubtiles[castPacket.SkillID]; ok {
		g.resolveSelfCenteredAoeHit(castPacket.SourceEntityID, castPacket.SkillID, radius)
		return
	}

	if castPacket.SkillID == d2hero.SkillChampStatique {
		g.resolveChampStatiqueHit(castPacket.SourceEntityID, castPacket.SkillID)
		return
	}

	if castPacket.SkillID == d2hero.SkillDistorsionTemporelle {
		g.resolveDistorsionTemporelleHit()
		return
	}

	if castPacket.SkillID == d2hero.SkillApocalypse {
		g.resolveApocalypseHit(castPacket.SourceEntityID, castPacket.SkillID)
		return
	}

	if castPacket.SkillID == d2hero.SkillBouclierDeMana {
		g.resolveBouclierDeManaHit(castPacket.SourceEntityID)
		return
	}

	if castPacket.SkillID == d2hero.SkillArmureDeGlace {
		g.resolveArmureDeGlaceHit(castPacket.SourceEntityID)
		return
	}

	if castPacket.SkillID == d2hero.SkillEveilDuNexus {
		g.resolveEveilDuNexusHit(castPacket.SourceEntityID)
		return
	}

	target := d2vector.NewPosition(castPacket.TargetX, castPacket.TargetY)

	if radius, ok := aoeAtTargetRadiusSubtiles[castPacket.SkillID]; ok {
		g.resolveAoeHit(target, castPacket.SourceEntityID, castPacket.SkillID, radius)
		return
	}

	if castPacket.SkillID == d2hero.SkillTelekinesie {
		g.resolveTelekinesieHit(castPacket.SourceEntityID, target)
		return
	}

	if castPacket.SkillID == d2hero.SkillRalentissement {
		g.resolveRalentissementHit(target)
		return
	}

	if castPacket.SkillID == d2hero.SkillTeleportation {
		g.resolveTeleportationHit(castPacket.SourceEntityID, target)
		return
	}

	if castPacket.SkillID == d2hero.SkillPrisonDeGlace {
		g.resolvePrisonDeGlaceHit(target)
		return
	}

	if castPacket.SkillID == d2hero.SkillVortex {
		g.resolveVortexHit(target)
		return
	}

	nearest := g.nearestKillableNPC(target, meleeHitRadiusSubtiles, nil)
	if nearest == nil {
		return
	}

	if castPacket.SkillID == d2hero.SkillEclairEnChaine {
		g.resolveChainHit(nearest, castPacket.SourceEntityID, castPacket.SkillID)
		return
	}

	if castPacket.SkillID == d2hero.SkillAmplification {
		until := g.clock().Add(amplificationDuration)
		nearest.ApplyAmplification(until)
		g.broadcastNPCStatusEffect(nearest, d2netpacket.NPCStatusAmplified, until)

		return
	}

	if castPacket.SkillID == d2hero.SkillRuptureArcane {
		until := g.clock().Add(ruptureArcaneDuration)
		nearest.ApplyResistanceStrip(until)
		g.broadcastNPCStatusEffect(nearest, d2netpacket.NPCStatusResistanceStripped, until)

		return
	}

	g.applyHit(nearest, castPacket.SourceEntityID, castPacket.SkillID)
}

// amplificationDuration is how long Amplification's debuff lasts on the
// NPC it's cast on.
const amplificationDuration = 4 * time.Second

// ruptureArcaneDuration is how long Rupture arcane's resistance removal
// lasts on the NPC it's cast on -- same placeholder duration as
// Amplification, the other tier-appropriate single-target Arcane debuff, in
// the absence of a design-specified number.
const ruptureArcaneDuration = 4 * time.Second

// resonanceMagiqueWindow is how long Résonance magique's "dégâts du
// suivant" bonus stays armed after a cast before going unused -- the
// design says "temporaire" but gives no number, same placeholder practice
// as amplificationDuration/ruptureArcaneDuration.
const resonanceMagiqueWindow = 3 * time.Second

// nearestKillableNPC returns the closest killable, living NPC to from
// within radiusSubtiles, skipping any NPC ID present in exclude (nil is a
// valid empty exclusion set). nil if none qualify.
func (g *GameServer) nearestKillableNPC(from d2vector.Position, radiusSubtiles float64, exclude map[string]bool) *d2mapentity.NPC {
	if len(g.mapEngines) == 0 {
		return nil
	}

	var nearest *d2mapentity.NPC

	nearestDist := math.MaxFloat64

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 || exclude[npc.ID()] {
			continue
		}

		pos := npc.GetPosition()
		if dist := from.Distance(&pos.Vector); dist < nearestDist {
			nearest, nearestDist = npc, dist
		}
	}

	if nearest == nil || nearestDist > radiusSubtiles {
		return nil
	}

	return nearest
}

// npcByID returns the live NPC entity with the given ID, or nil if none
// matches (including when there's no map engine at all). Unlike
// nearestKillableNPC, this doesn't filter by killable/HP -- callers that
// need a killable NPC should check IsKillable() themselves.
func (g *GameServer) npcByID(id string) *d2mapentity.NPC {
	if len(g.mapEngines) == 0 {
		return nil
	}

	for _, entity := range g.mapEngines[0].Entities() {
		if npc, ok := entity.(*d2mapentity.NPC); ok && npc.ID() == id {
			return npc
		}
	}

	return nil
}

// applyHit resolves a single hit against npc using resolveAttackDamage's
// Energy-scaled damage. Shared by resolveMeleeHit's single-target path and
// resolveChainHit.
func (g *GameServer) applyHit(npc *d2mapentity.NPC, sourceEntityID string, skillID int) {
	g.applyResolvedDamage(npc, sourceEntityID, skillID, g.resolveAttackDamage(sourceEntityID, skillID))
}

// amplificationDamagePercent is how much of its normal amount a hit deals
// against an amplified NPC (Amplification, "Augmente les dégâts magiques
// reçus par la cible" -- devil_game_design_reference.md §7). 150 = +50%.
//
// ponytail: the design names the effect but not a percentage; +50% is a
// placeholder pending real balance numbers.
const amplificationDamagePercent = 150

// amplifiedDamage scales damage by amplificationDamagePercent if amplified,
// otherwise returns it unchanged. Factored out of applyResolvedDamage so the
// scaling itself is testable without a live d2mapentity.NPC.
func amplifiedDamage(damage int, amplified bool) int {
	if !amplified {
		return damage
	}

	return damage * amplificationDamagePercent / 100
}

// applyResolvedDamage applies an already-computed damage amount to npc:
// Amplification's damage scaling, death/gold-award/XP-award, Éclat de
// glace's slow, and the NPCHit broadcast. Factored out of applyHit so
// skills whose damage isn't resolveAttackDamage's Energy-scaled formula --
// e.g. Champ statique's percent-of-current-HP -- can still share the
// death/broadcast plumbing.
func (g *GameServer) applyResolvedDamage(npc *d2mapentity.NPC, sourceEntityID string, skillID, damage int) {
	damage = amplifiedDamage(damage, npc.IsAmplified(g.clock()))

	resistance := npc.MagicResistancePercent()
	if npc.IsResistanceStripped(g.clock()) {
		resistance = 0 // Rupture arcane: "supprime les résistances d'une cible"
	}

	damage = d2hero.MitigateDamage(damage, resistance)

	died := npc.ApplyDamage(damage)

	if died {
		g.mapEngines[0].RemoveEntity(npc)
		g.awardGold(sourceEntityID)
		g.awardExperience(sourceEntityID)
		g.restoreManaOnKill(sourceEntityID)
	} else if skillID == d2hero.SkillEclatDeGlace {
		until := g.clock().Add(eclatDeGlaceSlowDuration)
		npc.ApplySlow(until)
		g.broadcastNPCStatusEffect(npc, d2netpacket.NPCStatusSlowed, until)
	}

	hitPacket, err := d2netpacket.CreateNPCHitPacket(npc.ID(), npc.HP, died)
	if err != nil {
		g.Errorf("CreateNPCHitPacket: %v", err)
		return
	}

	g.sendPacketToClients(hitPacket)
}

// resolveChainHit resolves Éclair en chaîne: hits first, then jumps to up
// to skillEclairEnChaineTargets-1 more nearby killable NPCs (each within
// eclairEnChaineChainRadius of the previous one), applying a hit to each.
//
// ponytail: no damage falloff between jumps -- each hit uses the same
// resolveAttackDamage as a single-target cast, and the design doesn't
// specify a falloff.
func (g *GameServer) resolveChainHit(first *d2mapentity.NPC, sourceEntityID string, skillID int) {
	hit := make(map[string]bool, skillEclairEnChaineTargets)
	current := first

	for i := 0; i < skillEclairEnChaineTargets && current != nil; i++ {
		g.applyHit(current, sourceEntityID, skillID)
		hit[current.ID()] = true

		current = g.nearestKillableNPC(current.GetPosition(), eclairEnChaineChainRadius, hit)
	}
}

// resolveSelfCenteredAoeHit hits every killable NPC within radiusSubtiles of
// sourceEntityID's own position (read off their HeroState -- the server has
// no map-entity for players, see nearestPlayer). Shared by any spell
// centered on the caster rather than a targeted position -- Nova de givre
// and Tempête de lames both funnel through this. A no-op if sourceEntityID
// isn't a resolved connected player.
//
// Limite de test connue (same as resolveChainHit): exercising the actual
// multi-NPC AoE path needs live d2mapentity.NPC instances in a map engine,
// whose fields are private outside their own package -- only the no-op path
// is unit-tested here.
func (g *GameServer) resolveSelfCenteredAoeHit(sourceEntityID string, skillID int, radiusSubtiles float64) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil {
		return
	}

	center := d2vector.NewPosition(state.X, state.Y)
	g.resolveAoeHit(center, sourceEntityID, skillID, radiusSubtiles)
}

// resolveAoeHit hits every killable NPC within radiusSubtiles of center.
// Shared by any Élémentalisme spell whose damage isn't a single/chain hit --
// Nova de givre (centered on the caster) and Boule de feu (centered on the
// cast's targeted position) both funnel through this.
func (g *GameServer) resolveAoeHit(center d2vector.Position, sourceEntityID string, skillID int, radiusSubtiles float64) {
	for _, npc := range g.killableNPCsWithin(center, radiusSubtiles) {
		g.applyHit(npc, sourceEntityID, skillID)
	}
}

// killableNPCsWithin returns every killable, living NPC within radiusSubtiles
// of center. Unlike nearestKillableNPC, this collects all matches rather
// than just the closest one -- for AoE spells like Nova de givre.
func (g *GameServer) killableNPCsWithin(center d2vector.Position, radiusSubtiles float64) []*d2mapentity.NPC {
	if len(g.mapEngines) == 0 {
		return nil
	}

	var found []*d2mapentity.NPC

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		pos := npc.GetPosition()
		if center.Distance(&pos.Vector) <= radiusSubtiles {
			found = append(found, npc)
		}
	}

	return found
}

// champStatiqueDamagePercent is the % of an NPC's own current HP that
// Champ statique removes from every killable NPC on the map
// (devil_game_design_reference.md §7 Arcane: "Réduit la vie de toutes les
// entités à l'écran d'un % fixe").
//
// ponytail: "toutes les entités à l'écran" is modeled as every killable NPC
// on the map, not an actual per-client screen/viewport query -- the server
// has no concept of what's currently visible to a given client. Upgrade
// path: filter killableNPCsWithin by a real screen-radius constant, once
// one exists, instead of scanning every NPC.
const champStatiqueDamagePercent = 20

// resolveChampStatiqueHit applies champStatiqueDamagePercent of current HP
// to every killable NPC on the map, via applyResolvedDamage (so each still
// dies/awards gold+XP/broadcasts normally -- only the damage source
// differs from resolveAttackDamage's Energy scaling).
func (g *GameServer) resolveChampStatiqueHit(sourceEntityID string, skillID int) {
	if len(g.mapEngines) == 0 {
		return
	}

	state := g.playerStateOf(sourceEntityID)

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		damage := npc.HP * champStatiqueDamagePercent / 100

		// Correction (août 2026): Champ statique bypasses resolveAttackDamage
		// entirely (it deals a % of current HP, not the base_sort/Energy
		// formula), so it never consumed Résonance magique's bonus even
		// though it's squarely an Arcane skill the synergy is meant to
		// reach. state can be nil here (unresolved caster) where
		// resolveAttackDamage would already have bailed out earlier.
		if state != nil && state.Stats != nil {
			if percent := g.resonanceMagiqueBonusPercent(state, skillID); percent > 0 {
				damage += (damage * percent) / 100
			}
		}

		g.applyResolvedDamage(npc, sourceEntityID, skillID, damage)
	}
}

// distorsionTemporelleSlowDuration is exactly what the design specifies
// (devil_game_design_reference.md §7 Arcane: "Ralentit toutes les entités à
// l'écran pendant 5 secondes") -- unlike Ralentissement's own placeholder
// duration, no guesswork needed here.
const distorsionTemporelleSlowDuration = 5 * time.Second

// resolveDistorsionTemporelleHit slows (d2mapentity.NPC.ApplySlow) every
// killable NPC on the map, same "toutes les entités à l'écran" modeling
// choice as Champ statique (every NPC on the map, not a real per-client
// screen/viewport query -- see champStatiqueDamagePercent's doc comment).
// Deals no damage.
func (g *GameServer) resolveDistorsionTemporelleHit() {
	if len(g.mapEngines) == 0 {
		return
	}

	until := g.clock().Add(distorsionTemporelleSlowDuration)

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		npc.ApplySlow(until)
		g.broadcastNPCStatusEffect(npc, d2netpacket.NPCStatusSlowed, until)
	}
}

// resolveApocalypseHit hits every killable NPC on the map via applyHit
// (resolveAttackDamage's normal Energy-scaled damage) -- Élémentalisme's
// ultimate, same "toute la zone visible" modeling choice as
// resolveDistorsionTemporelleHit/resolveChampStatiqueHit.
func (g *GameServer) resolveApocalypseHit(sourceEntityID string, skillID int) {
	if len(g.mapEngines) == 0 {
		return
	}

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		g.applyHit(npc, sourceEntityID, skillID)
	}
}

// broadcastPlayerStatusEffect tells clients about a change to playerID's
// own status effect -- used by Bouclier de mana/Armure de glace's toggles
// and Éveil du Nexus's timed immunity, none of which broadcast anything
// before this. Even solo play round-trips through a local client/server
// (ROADMAP.md), so without this the client's own copy of its player never
// found out, unlike every other Devil mechanic this session.
func (g *GameServer) broadcastPlayerStatusEffect(playerID, effect string, active bool, until time.Time) {
	statusPacket, err := d2netpacket.CreatePlayerStatusEffectPacket(playerID, effect, active, until)
	if err != nil {
		g.Errorf("CreatePlayerStatusEffectPacket: %v", err)
		return
	}

	g.sendPacketToClients(statusPacket)
}

// resolveBouclierDeManaHit toggles sourceEntityID's own
// HeroStatsState.ManaShieldActive. The absorption itself
// (ApplyDamageWithManaShield) is already wired into tryMonsterAttack from
// an earlier phase -- this is just the first real trigger a player has to
// turn it on and off, rather than it being set directly for tests only.
// A no-op if sourceEntityID isn't a resolved connected player with stats.
func (g *GameServer) resolveBouclierDeManaHit(sourceEntityID string) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	state.Stats.ManaShieldActive = !state.Stats.ManaShieldActive
	g.broadcastPlayerStatusEffect(sourceEntityID, d2netpacket.PlayerStatusManaShield, state.Stats.ManaShieldActive, time.Time{})
}

// resolveArmureDeGlaceHit toggles sourceEntityID's own
// HeroStatsState.ArmureDeGlaceActive. The damage reduction and attacker slow
// it enables are applied in tryMonsterAttack. A no-op if sourceEntityID
// isn't a resolved connected player with stats.
func (g *GameServer) resolveArmureDeGlaceHit(sourceEntityID string) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	state.Stats.ArmureDeGlaceActive = !state.Stats.ArmureDeGlaceActive
	g.broadcastPlayerStatusEffect(sourceEntityID, d2netpacket.PlayerStatusArmureDeGlace, state.Stats.ArmureDeGlaceActive, time.Time{})
}

// eveilDuNexusImmunityDuration is exactly what the design specifies
// (devil_game_design_reference.md §7 Ésotérisme: "immunité magique pendant
// 8 secondes") -- no guesswork needed, same as Distorsion temporelle's own
// exact duration.
const eveilDuNexusImmunityDuration = 8 * time.Second

// eveilDuNexusHealPercent is how much of MaxHealth Éveil du Nexus restores
// ("soigne le Mage").
//
// ponytail: the design names the heal but not an amount; 30% of MaxHealth
// is a placeholder pending real balance numbers.
const eveilDuNexusHealPercent = 30

// resolveEveilDuNexusHit grants sourceEntityID's own HeroStatsState magic
// immunity for eveilDuNexusImmunityDuration and heals them by
// eveilDuNexusHealPercent of MaxHealth. A no-op if sourceEntityID isn't a
// resolved connected player with stats.
func (g *GameServer) resolveEveilDuNexusHit(sourceEntityID string) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return
	}

	until := g.clock().Add(eveilDuNexusImmunityDuration)
	state.Stats.ApplyMagicImmunity(until)
	g.broadcastPlayerStatusEffect(sourceEntityID, d2netpacket.PlayerStatusMagicImmune, true, until)

	state.Stats.Heal(state.Stats.MaxHealth * eveilDuNexusHealPercent / 100)

	// Correction (août 2026): the heal itself was never broadcast either --
	// reuses PlayerDamagedPacket (it just carries the player's current HP,
	// not a damage delta, so it doubles fine for a heal) rather than adding
	// a redundant packet type.
	healedPacket, err := d2netpacket.CreatePlayerDamagedPacket(sourceEntityID, state.Stats.Health, false)
	if err != nil {
		g.Errorf("CreatePlayerDamagedPacket: %v", err)
		return
	}

	g.sendPacketToClients(healedPacket)
}

// broadcastNPCMoved tells clients about npc's current position -- used
// after a server-authoritative instant displacement (Knockback/Pull) that
// would otherwise be invisible to them, since neither mutates HP (the only
// other thing NPCHit already broadcasts).
func (g *GameServer) broadcastNPCMoved(npc *d2mapentity.NPC) {
	pos := npc.GetPosition()

	movedPacket, err := d2netpacket.CreateNPCMovedPacket(npc.ID(), pos.X(), pos.Y())
	if err != nil {
		g.Errorf("CreateNPCMovedPacket: %v", err)
		return
	}

	g.sendPacketToClients(movedPacket)
}

// broadcastNPCStatusEffect tells clients that npc has effect applied until
// the given time -- Éclat de glace/Ralentissement/Distorsion temporelle's
// slow, Prison de glace's immobilize, Amplification's damage amplification,
// Rupture arcane's resistance strip. Each of these previously only mutated
// the server's own copy of the NPC, with clients never finding out.
func (g *GameServer) broadcastNPCStatusEffect(npc *d2mapentity.NPC, effect string, until time.Time) {
	statusPacket, err := d2netpacket.CreateNPCStatusEffectPacket(npc.ID(), effect, until)
	if err != nil {
		g.Errorf("CreateNPCStatusEffectPacket: %v", err)
		return
	}

	g.sendPacketToClients(statusPacket)
}

// resolveTelekinesieHit knocks every killable NPC within telekinesieRadiusSubtiles
// of target away from sourceEntityID's own position, via d2mapentity.NPC.Knockback.
// A no-op (besides the position scan) if sourceEntityID isn't a resolved
// connected player.
//
// ponytail: deals no damage: only the knockback half of "Repousse les
// entités, interaction avec les objets à distance" is modeled -- the object-
// interaction half needs Phase 5's item system.
func (g *GameServer) resolveTelekinesieHit(sourceEntityID string, target d2vector.Position) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil {
		return
	}

	casterPos := d2vector.NewPosition(state.X, state.Y)

	for _, npc := range g.killableNPCsWithin(target, telekinesieRadiusSubtiles) {
		npc.Knockback(casterPos, telekinesieKnockbackDistance)
		g.broadcastNPCMoved(npc)
	}
}

// ralentissementRadiusSubtiles/ralentissementSlowDuration: Ralentissement
// slows every killable NPC within this many subtiles of the cast's targeted
// position, for this long. The slow amount itself
// (npcSlowedSpeedMultiplier, applied by ChasePlayer) already matches the
// design's "réduit la vitesse des entités de 50%" -- same mechanic Éclat de
// glace uses on a single target, just over an area here.
const (
	ralentissementRadiusSubtiles = 4
	ralentissementSlowDuration   = 3 * time.Second
)

// resolveRalentissementHit applies a slow (d2mapentity.NPC.ApplySlow) to
// every killable NPC within ralentissementRadiusSubtiles of target. Deals no
// damage, so it doesn't go through applyResolvedDamage/broadcast an NPCHit --
// same documented client-sync gap as Éclat de glace's own slow (no packet
// carries SlowedUntil to clients yet).
func (g *GameServer) resolveRalentissementHit(target d2vector.Position) {
	until := g.clock().Add(ralentissementSlowDuration)

	for _, npc := range g.killableNPCsWithin(target, ralentissementRadiusSubtiles) {
		npc.ApplySlow(until)
		g.broadcastNPCStatusEffect(npc, d2netpacket.NPCStatusSlowed, until)
	}
}

// prisonDeGlaceRadiusSubtiles/prisonDeGlaceDuration: Prison de glace
// immobilizes every killable NPC within this many subtiles of the cast's
// targeted position, for this long.
const (
	prisonDeGlaceRadiusSubtiles = 4
	prisonDeGlaceDuration       = 3 * time.Second
)

// resolvePrisonDeGlaceHit applies a full immobilization
// (d2mapentity.NPC.ApplyImmobilize) to every killable NPC within
// prisonDeGlaceRadiusSubtiles of target. Deals no damage, so -- same as
// resolveRalentissementHit -- it doesn't go through
// applyResolvedDamage/broadcast an NPCHit; no packet carries
// ImmobilizedUntil to clients yet either.
func (g *GameServer) resolvePrisonDeGlaceHit(target d2vector.Position) {
	until := g.clock().Add(prisonDeGlaceDuration)

	for _, npc := range g.killableNPCsWithin(target, prisonDeGlaceRadiusSubtiles) {
		npc.ApplyImmobilize(until)
		g.broadcastNPCStatusEffect(npc, d2netpacket.NPCStatusImmobilized, until)
	}
}

// vortexRadiusSubtiles/vortexPullDistance: Vortex pulls every killable NPC
// within this many subtiles of the cast's targeted position that many
// subtiles closer to it (devil_game_design_reference.md §7 Arcane: "Aspire
// toutes les entités proches vers un point").
const (
	vortexRadiusSubtiles = 6
	vortexPullDistance   = 4
)

// resolveVortexHit pulls (d2mapentity.NPC.Pull) every killable NPC within
// vortexRadiusSubtiles of target that many subtiles closer to it. Deals no
// damage.
func (g *GameServer) resolveVortexHit(target d2vector.Position) {
	for _, npc := range g.killableNPCsWithin(target, vortexRadiusSubtiles) {
		npc.Pull(target, vortexPullDistance)
		g.broadcastNPCMoved(npc)
	}
}

// resolveTeleportationHit instantly moves sourceEntityID's own HeroState to
// target and broadcasts a PlayerTeleportedPacket so clients snap the
// player there (unlike MovePlayerPacket, which paths a smooth walk). A
// no-op if sourceEntityID isn't a resolved connected player.
func (g *GameServer) resolveTeleportationHit(sourceEntityID string, target d2vector.Position) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil {
		return
	}

	state.X, state.Y = target.X(), target.Y()

	packet, err := d2netpacket.CreatePlayerTeleportedPacket(sourceEntityID, state.X, state.Y)
	if err != nil {
		g.Errorf("CreatePlayerTeleportedPacket: %v", err)
		return
	}

	g.sendPacketToClients(packet)
}

// resolveAttackDamage returns the damage a cast of skillID from
// sourceEntityID deals, per the design's magic damage formula
// (devil_game_design_reference.md §6): base spell damage scaled by the
// caster's Energy --
//
//	Dégâts = base_sort × (1 + Energy / 100)
//
// base_sort comes from d2hero.SkillBaseSortDamage(skillID, ...) -- see d2hero.DevilSkills.
// Falls back to the unscaled base damage if the attacker isn't a connected
// player or has no stats resolved.
//
// ponytail: no attack rating. See ROADMAP.md Phase 1/2 for the rest of the
// combat formula.
func (g *GameServer) resolveAttackDamage(sourceEntityID string, skillID int) int {
	baseSort := d2hero.SkillBaseSortDamage(skillID, baseSortDamage)

	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return baseSort
	}

	damage := baseSort + (baseSort*effectiveEnergy(state))/100

	// Correction (août 2026): this used to apply the equipped weapon's Fire
	// modifier to every skill regardless of its actual element -- a Fire%
	// staff was silently buffing Éclat de glace/Éclair en chaîne too. Now
	// gated on d2hero.DevilSkillDef.DealsFireDamage.
	if def, ok := d2hero.DevilSkills[skillID]; ok && def.DealsFireDamage {
		if firePercent := d2hero.ItemFireDamagePercent(state.Equipment.RightHand.GetItemCode()); firePercent > 0 {
			damage += (damage * firePercent) / 100
		}
	}

	// Maîtrise élémentaire only boosts Élémentalisme's own skills.
	if def, ok := d2hero.DevilSkills[skillID]; ok && def.Tree == d2hero.TreeElementalisme {
		if maitrise, learned := state.Skills[d2hero.SkillMaitriseElementaire]; learned {
			if percent := d2hero.MaitriseElementaireDamagePercent(maitrise.SkillPoints); percent > 0 {
				damage += (damage * percent) / 100
			}
		}
	}

	if percent := g.resonanceMagiqueBonusPercent(state, skillID); percent > 0 {
		damage += (damage * percent) / 100
	}

	return damage
}

// resonanceMagiqueBonusPercent returns the damage bonus percent Résonance
// magique grants this cast of skillID, consuming the caster's armed bonus
// (HeroStatsState.ConsumeResonanceMagiqueBonus) in the process. 0 (and the
// armed bonus left untouched) if skillID isn't Élémentalisme/Arcane
// ("tous les sorts actifs des arbres I et II", §7 "Synergies") or the
// caster hasn't learned the passive -- checked first so a caster who never
// took it never touches (or clears) the armed timer. Shared by
// resolveAttackDamage and resolveChampStatiqueHit, the only two damage
// paths a Devil skill can take.
func (g *GameServer) resonanceMagiqueBonusPercent(state *d2hero.HeroState, skillID int) int {
	def, ok := d2hero.DevilSkills[skillID]
	if !ok || (def.Tree != d2hero.TreeElementalisme && def.Tree != d2hero.TreeArcane) {
		return 0
	}

	resonance, learned := state.Skills[d2hero.SkillResonanceMagique]
	if !learned || !state.Stats.ConsumeResonanceMagiqueBonus(g.clock()) {
		return 0
	}

	return d2hero.ResonanceMagiqueDamagePercent(resonance.SkillPoints)
}

var (
	errPlayerAlreadyExists = errors.New("player already exists")
	errServerFull          = errors.New("server full") // Server currently at maximum TCP connections
)

// GameServer manages a copy of the map and entities as well as manages packet routing and connections.
// It can accept connections from localhost as well remote clients. It can also be started in a standalone mode.
type GameServer struct {
	sync.RWMutex
	connections         map[string]ClientConnection
	listener            net.Listener
	networkServer       bool
	ctx                 context.Context
	cancel              context.CancelFunc
	asset               *d2asset.AssetManager
	mapEngines          []*d2mapengine.MapEngine
	scriptEngine        *d2script.ScriptEngine
	seed                int64
	maxConnections      int
	packetManagerChan   chan ReceivedPacket
	heroStateFactory    *d2hero.HeroStateFactory
	lastCastAt          map[string]time.Time
	lastMonsterAttackAt map[string]time.Time
	lastTranscendanceAt map[string]time.Time
	activeEventUntil    map[string]time.Time // see game_server_events.go
	clock               func() time.Time     // overridden in tests; defaults to time.Now

	*d2util.Logger
}

// ReceivedPacket encapsulates the data necessary for the packet manager goroutine to process data from clients.
// The packet manager needs to know who sent the data, in addition to the data itself.
type ReceivedPacket struct {
	Client ClientConnection
	Packet d2netpacket.NetPacket
}

// NewGameServer builds a new GameServer that can be started
//
// ctx: required context item
// networkServer: true = 0.0.0.0 | false = 127.0.0.1
// maxConnections (default: 8): maximum number of TCP connections allowed open
func NewGameServer(asset *d2asset.AssetManager,
	networkServer bool,
	l d2util.LogLevel,
	maxConnections ...int) (*GameServer,
	error) {
	if len(maxConnections) == 0 {
		maxConnections = []int{8}
	}

	heroStateFactory, err := d2hero.NewHeroStateFactory(asset)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	gameServer := &GameServer{
		ctx:                 ctx,
		cancel:              cancel,
		asset:               asset,
		connections:         make(map[string]ClientConnection),
		networkServer:       networkServer,
		maxConnections:      maxConnections[0],
		packetManagerChan:   make(chan ReceivedPacket),
		mapEngines:          make([]*d2mapengine.MapEngine, 0),
		scriptEngine:        d2script.CreateScriptEngine(),
		seed:                time.Now().UnixNano(),
		heroStateFactory:    heroStateFactory,
		lastCastAt:          make(map[string]time.Time),
		lastMonsterAttackAt: make(map[string]time.Time),
		lastTranscendanceAt: make(map[string]time.Time),
		activeEventUntil:    make(map[string]time.Time),
		clock:               time.Now,
	}

	gameServer.Logger = d2util.NewLogger()
	gameServer.Logger.SetPrefix(logPrefix)
	gameServer.Logger.SetLevel(l)

	mapEngine := d2mapengine.CreateMapEngine(l, asset)
	mapEngine.SetSeed(gameServer.seed)
	mapEngine.ResetMap(d2enum.RegionAct1Town, 100, 100)

	mapGen, err := d2mapgen.NewMapGenerator(asset, l, mapEngine)
	if err != nil {
		return nil, err
	}

	mapGen.GenerateAct1Overworld()

	gameServer.mapEngines = append(gameServer.mapEngines, mapEngine)

	// ponytail: exposes just a count, not the map engines themselves --
	// unlike otto, a WASM guest can't hold a live reference to a Go object
	// graph. The old getMapEngines hook returned the whole []*MapEngine
	// slice, which doesn't translate; nothing exercises this today, so it's
	// a placeholder proving the host-function mechanism works, not a
	// designed script API. See ROADMAP.md Phase 3/2 for the real one.
	gameServer.scriptEngine.AddFunction("mapEngineCount", func() uint32 {
		return uint32(len(gameServer.mapEngines))
	})

	return gameServer, nil
}

// Start essentially starts all of the game server go routines as well as begins listening for connection. This will
// return an error if it is unable to bind to a socket.
func (g *GameServer) Start() error {
	listenerAddress := "127.0.0.1:" + port
	if g.networkServer {
		listenerAddress = "0.0.0.0:" + port
	}

	g.Infof("Starting Game Server @ %s\n", listenerAddress)

	l, err := net.Listen("tcp4", listenerAddress)
	if err != nil {
		return err
	}

	g.listener = l

	go g.packetManager()
	go g.runMonsterAILoop()

	go func() {
		for {
			c, err := g.listener.Accept()
			if err != nil {
				select {
				case <-g.ctx.Done():
					// this error was just a result of the server closing, don't worry about it
				default:
					g.Errorf("Unable to accept connection: %s", err)
				}

				return
			}

			go g.handleConnection(c)
		}
	}()

	return nil
}

// Stop stops the game server
func (g *GameServer) Stop() {
	g.Lock()
	g.cancel()
	g.connections = make(map[string]ClientConnection)

	if err := g.listener.Close(); err != nil {
		g.Errorf("failed to close the listener %s, err: %v\n", g.listener.Addr(), err)
	}
}

// packetManager is meant to be started as a Goroutine and is used to manage routing of packets to clients.
func (g *GameServer) packetManager() {
	defer close(g.packetManagerChan)

	for {
		select {
		// If the server is stopped we need to clean up the packet manager goroutine
		case <-g.ctx.Done():
			return
		case p := <-g.packetManagerChan:
			err := g.OnPacketReceived(p.Client, p.Packet)
			if err != nil {
				g.Errorf("failed to handle packet received from client %s: %v", p.Client.GetUniqueID(), err)
			}
		}
	}
}

func (g *GameServer) sendPacketToClients(packet d2netpacket.NetPacket) {
	for _, c := range g.connections {
		if err := c.SendPacketToClient(packet); err != nil {
			g.Errorf("GameServer: error sending packet: %s to client %s: %s", packet.PacketType, c.GetUniqueID(), err)
		}
	}
}

// handleConnection accepts an individual connection and starts pooling for new packets. It is recommended this is called
// via Go Routine. Context should be a property of the GameServer Struct.
func (g *GameServer) handleConnection(conn net.Conn) {
	var (
		connected int
		client    ClientConnection
	)

	g.Infof("Accepting connection: %s\n", conn.RemoteAddr().String())

	defer func() {
		if err := conn.Close(); err != nil {
			g.Errorf("failed to close the connection: %s\n", conn.RemoteAddr())
		}
	}()

	decoder := json.NewDecoder(conn)

	for {
		var packet d2netpacket.NetPacket

		err := decoder.Decode(&packet)
		if err != nil {
			switch err {
			case io.EOF:
				break // the other side closed the connection
			default:
				g.Error(err.Error())
			}

			return // allow the connection to close
		}

		// If this is the first packet we are seeing from this specific connection we first need to see if the client
		// is sending a valid request. If this is a valid request, we will register it and flip the connected switch
		// to.
		if connected == 0 {
			if packet.PacketType != d2netpackettype.PlayerConnectionRequest {
				g.Infof("Closing connection with %s: did not receive new player connection request...", conn.RemoteAddr().String())
			}

			if client, err = g.registerConnection(packet.PacketData, conn); err != nil {
				return
			}

			connected = 1
		}

		select {
		case <-g.ctx.Done():
			return
		default:
			g.packetManagerChan <- ReceivedPacket{
				Client: client,
				Packet: packet,
			}
		}
	}
}

// registerConnection accepts a PlayerConnectionRequestPacket and thread safely updates the connection pool
//
// Errors:
// - errServerFull
// - errPlayerAlreadyExists
func (g *GameServer) registerConnection(b []byte, conn net.Conn) (ClientConnection, error) {
	var client ClientConnection

	g.Lock()
	defer g.Unlock()

	// check to see if the server is full
	if len(g.connections) >= g.maxConnections {
		sf, serverFullErr := d2netpacket.CreateServerFullPacket()
		if serverFullErr != nil {
			g.Errorf("ServerFullPacket: %v", serverFullErr)
		}

		msf, marshalServerFullErr := d2netpacket.MarshalPacket(sf)
		if marshalServerFullErr != nil {
			g.Errorf("MarshalPacket: %v", marshalServerFullErr)
		}

		_, errServerFullPacket := conn.Write(msf)
		g.Warningf("%v", errServerFullPacket)

		return client, errServerFull
	}

	// if it is not full, unmarshal the playerConnectionRequest
	packet, err := d2netpacket.UnmarshalPlayerConnectionRequest(b)
	if err != nil {
		g.Errorf("Failed to unmarshal PlayerConnectionRequest: %s\n", err)
	}

	// check to see if the player is already registered
	if _, ok := g.connections[packet.ID]; ok {
		g.Errorf("%v", errPlayerAlreadyExists)
		return client, errPlayerAlreadyExists
	}

	// Client a new TCP Client Connection and add it to the connections map
	client = d2tcpclientconnection.CreateTCPClientConnection(conn, packet.ID)
	client.SetPlayerState(packet.PlayerState)

	g.OnClientConnected(client)

	return client, nil
}

// OnClientConnected initializes the given ClientConnection. It sends the
// following packets to the newly connected client: UpdateServerInfoPacket,
// GenerateMapPacket, AddPlayerPacket.
//
// It also sends AddPlayerPackets for each other player entity to the new
// player and vice versa, so all player entities exist on all clients.
//
// For more information, see d2networking.d2netpacket.
func (g *GameServer) OnClientConnected(client ClientConnection) {
	// Temporary position hack --------------------------------------------
	// https://github.com/OpenDiablo2/OpenDiablo2/issues/829
	sx, sy := g.mapEngines[0].GetStartPosition()
	clientPlayerState := client.GetPlayerState()
	clientPlayerState.X = sx
	clientPlayerState.Y = sy
	// --------------------------------------------------------------------

	g.Infof("Client connected with an id of %s", client.GetUniqueID())
	g.connections[client.GetUniqueID()] = client

	g.handleClientConnection(client, sx, sy)
}

func (g *GameServer) handleClientConnection(client ClientConnection, x, y float64) {
	usi, err := d2netpacket.CreateUpdateServerInfoPacket(g.seed, client.GetUniqueID())
	if err != nil {
		g.Errorf("UpdateServerInfoPacket: %v", err)
	}

	err = client.SendPacketToClient(usi)
	if err != nil {
		g.Errorf("GameServer: error sending UpdateServerInfoPacket to client %s: %s", client.GetUniqueID(), err)
	}

	gmp, err := d2netpacket.CreateGenerateMapPacket(d2enum.RegionAct1Town)
	if err != nil {
		g.Errorf("GenerateMapPacket: %v", err)
	}

	err = client.SendPacketToClient(gmp)
	if err != nil {
		g.Errorf("GameServer: error sending GenerateMapPacket to client %s: %s", client.GetUniqueID(), err)
	}

	playerState := client.GetPlayerState()

	// these are in subtiles
	playerX := int(x*subtilesPerTile) + middleOfTileOffset
	playerY := int(y*subtilesPerTile) + middleOfTileOffset

	d2hero.HydrateSkills(playerState.Skills, g.asset)

	createPlayerPacket, err := d2netpacket.CreateAddPlayerPacket(
		client.GetUniqueID(),
		playerState.HeroName,
		playerX,
		playerY,
		playerState.HeroType,
		playerState.Stats,
		playerState.Skills,
		playerState.Equipment,
		playerState.LeftSkill,
		playerState.RightSkill,
		playerState.Gold,
	)
	if err != nil {
		g.Errorf("AddPlayerPacket: %v", err)
	}

	for _, connection := range g.connections {
		err := connection.SendPacketToClient(createPlayerPacket)
		if err != nil {
			g.Errorf("GameServer: error sending %T to client %s: %s", createPlayerPacket, connection.GetUniqueID(), err)
		}

		if connection.GetUniqueID() == client.GetUniqueID() {
			continue
		}

		conPlayerState := connection.GetPlayerState()
		playerX := int(conPlayerState.X*subtilesPerTile) + middleOfTileOffset
		playerY := int(conPlayerState.Y*subtilesPerTile) + middleOfTileOffset
		app, err := d2netpacket.CreateAddPlayerPacket(
			connection.GetUniqueID(),
			conPlayerState.HeroName,
			playerX,
			playerY,
			conPlayerState.HeroType,
			conPlayerState.Stats,
			conPlayerState.Skills,
			conPlayerState.Equipment,
			conPlayerState.LeftSkill,
			conPlayerState.RightSkill,
			conPlayerState.Gold,
		)

		if err != nil {
			g.Errorf("AddPlayerPacket: %v", err)
		}

		err = client.SendPacketToClient(app)

		if err != nil {
			g.Errorf("GameServer: error sending CreateAddPlayerPacket to client %s: %s", connection.GetUniqueID(), err)
		}
	}
}

// OnClientDisconnected removes the given client from the list
// of client connections.
// If this client was the host, disconnects all clients and kills GameServer.
func (g *GameServer) OnClientDisconnected(client ClientConnection) {
	g.Infof("Client disconnected with an id of %s", client.GetUniqueID())
	delete(g.connections, client.GetUniqueID())

	if client.GetConnectionType() == d2clientconnectiontype.Local {
		g.Info("Host disconnected, game server shuting down")

		serverClosed, err := d2netpacket.CreateServerClosedPacket()
		if err != nil {
			g.Errorf("failed to generate ServerClosed packet after host disconnected: %s", err)
		} else {
			g.sendPacketToClients(serverClosed)
		}

		g.Stop()
	}
}

// OnPacketReceived is called when a packet has been received from a remote client,
// and by the local client to 'send' a packet to the server,
// nolint:gocyclo // switch statement on packet type makes sense, no need to change
func (g *GameServer) OnPacketReceived(client ClientConnection, packet d2netpacket.NetPacket) error {
	if g == nil {
		return errors.New("game server is nil")
	}

	switch packet.PacketType {
	case d2netpackettype.MovePlayer:
		movePacket, err := d2netpacket.UnmarshalMovePlayer(packet.PacketData)
		if err != nil {
			return err
		}

		playerState := g.connections[client.GetUniqueID()].GetPlayerState()
		playerState.X = movePacket.DestX
		playerState.Y = movePacket.DestY

		g.sendPacketToClients(packet)
	case d2netpackettype.CastSkill:
		g.resolveMeleeHit(packet)
		g.sendPacketToClients(packet)
	case d2netpackettype.UsePotionRequest:
		g.resolveUsePotion(packet)
	case d2netpackettype.LearnSkillRequest:
		g.resolveLearnSkill(packet)
	case d2netpackettype.RespecSkillsRequest:
		g.resolveRespecSkills(packet)
	case d2netpackettype.RespecSingleSkillRequest:
		g.resolveRespecSingleSkill(packet)
	case d2netpackettype.InvestSkillPointRequest:
		g.resolveInvestSkillPoint(packet)
	case d2netpackettype.SpendAttributePointRequest:
		g.resolveSpendAttributePoint(packet)
	case d2netpackettype.RespecSingleAttributePointRequest:
		g.resolveRespecSingleAttributePoint(packet)
	case d2netpackettype.SpawnItem:
		g.sendPacketToClients(packet)
	case d2netpackettype.SavePlayer:
		savePacket, err := d2netpacket.UnmarshalSavePlayer(packet.PacketData)
		if err != nil {
			return err
		}

		playerState := g.connections[client.GetUniqueID()].GetPlayerState()
		playerState.LeftSkill = savePacket.Player.LeftSkill.Shallow.SkillID
		playerState.RightSkill = savePacket.Player.RightSkill.Shallow.SkillID
		playerState.Stats = savePacket.Player.Stats
		playerState.Act = savePacket.Player.Act
		playerState.Difficulty = savePacket.Difficulty

		err = g.heroStateFactory.Save(playerState)
		if err != nil {
			g.Errorf("GameServer: error saving saving Player: %s", err)
		}
	case d2netpackettype.PlayerConnectionRequest:
		break // prevent log message. these are handled by handleConnection
	case d2netpackettype.PlayerDisconnectionNotification:
		g.sendPacketToClients(packet)
		g.OnClientDisconnected(client)
	default:
		g.Warningf("GameServer: received unknown packet %s", packet.PacketType)
	}

	return nil
}
