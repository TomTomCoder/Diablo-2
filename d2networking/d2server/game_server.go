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
			regen := int(manaRegenPerSecond(state.Stats.Energy) * now.Sub(last).Seconds())
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

// dexterityOf returns sourceEntityID's Dexterity, or 0 if it isn't a
// connected player or has no stats resolved.
func (g *GameServer) dexterityOf(sourceEntityID string) int {
	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return 0
	}

	return state.Stats.Dexterity
}

// aiTickInterval is how often the monster AI loop reevaluates.
const aiTickInterval = 200 * time.Millisecond

// ponytail: flat numbers for every monster (aggro/attack range, damage,
// attack cooldown) instead of per-monster data -- no monster combat data
// exists yet (ROADMAP.md Phase 4). Same placeholder shape as the player's
// own combat constants above.
const (
	monsterAggroRadiusSubtiles = 8
	monsterAttackRangeSubtiles = 2
	monsterAttackCooldown      = 1200 * time.Millisecond
	monsterAttackDamage        = 3
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

// novaDeGivreRadiusSubtiles: Nova de givre hits every killable NPC within
// this many subtiles of the caster (devil_game_design_reference.md §7:
// "Explosion de froid en zone autour du Mage") -- centered on the caster,
// not on the cast's targeted position like every other Élémentalisme spell.
const novaDeGivreRadiusSubtiles = 5

// bouleDeFeuRadiusSubtiles/tempeteStatiqueRadiusSubtiles: how many subtiles
// around the cast's targeted position these AoE-at-target spells hit
// ("Projectile AoE, dégâts feu élevés" / "Invoque une zone d'éclair
// persistante" -- see SkillTempeteStatique's doc comment for why the
// "persistante" part isn't modeled yet).
const (
	bouleDeFeuRadiusSubtiles      = 3
	tempeteStatiqueRadiusSubtiles = 4
)

// aoeAtTargetRadiusSubtiles maps each AoE-at-target-position Élémentalisme
// skill to its radius, so resolveMeleeHit can dispatch to resolveAoeHit
// without a growing if-chain as more such spells are added (Orbe glaciale,
// Météore, Apocalypse are all this shape too -- ROADMAP.md Phase 2).
var aoeAtTargetRadiusSubtiles = map[int]float64{
	d2hero.SkillBouleDeFeu:      bouleDeFeuRadiusSubtiles,
	d2hero.SkillTempeteStatique: tempeteStatiqueRadiusSubtiles,
}

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

		if dist > monsterAggroRadiusSubtiles {
			continue
		}

		if dist > monsterAttackRangeSubtiles {
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

	g.Lock()
	last, attacked := g.lastMonsterAttackAt[npcID]

	if attacked && now.Sub(last) < monsterAttackCooldown {
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
	damage := d2hero.MitigateDamage(monsterAttackDamage, d2hero.CapResistance(state.Stats.FireResist))
	died := state.Stats.ApplyDamageWithManaShield(damage)

	packet, err := d2netpacket.CreatePlayerDamagedPacket(playerID, state.Stats.Health, died)
	if err != nil {
		g.Errorf("CreatePlayerDamagedPacket: %v", err)
		return
	}

	g.sendPacketToClients(packet)
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

	if castPacket.SkillID == d2hero.SkillNovaDeGivre {
		g.resolveNovaHit(castPacket.SourceEntityID, castPacket.SkillID)
		return
	}

	if castPacket.SkillID == d2hero.SkillChampStatique {
		g.resolveChampStatiqueHit(castPacket.SourceEntityID, castPacket.SkillID)
		return
	}

	target := d2vector.NewPosition(castPacket.TargetX, castPacket.TargetY)

	if radius, ok := aoeAtTargetRadiusSubtiles[castPacket.SkillID]; ok {
		g.resolveAoeHit(target, castPacket.SourceEntityID, castPacket.SkillID, radius)
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

	g.applyHit(nearest, castPacket.SourceEntityID, castPacket.SkillID)
}

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

// applyHit resolves a single hit against npc using resolveAttackDamage's
// Energy-scaled damage. Shared by resolveMeleeHit's single-target path and
// resolveChainHit.
func (g *GameServer) applyHit(npc *d2mapentity.NPC, sourceEntityID string, skillID int) {
	g.applyResolvedDamage(npc, sourceEntityID, skillID, g.resolveAttackDamage(sourceEntityID, skillID))
}

// applyResolvedDamage applies an already-computed damage amount to npc:
// death/gold-award/XP-award, Éclat de glace's slow, and the NPCHit
// broadcast. Factored out of applyHit so skills whose damage isn't
// resolveAttackDamage's Energy-scaled formula -- e.g. Champ statique's
// percent-of-current-HP -- can still share the death/broadcast plumbing.
func (g *GameServer) applyResolvedDamage(npc *d2mapentity.NPC, sourceEntityID string, skillID, damage int) {
	died := npc.ApplyDamage(damage)

	if died {
		g.mapEngines[0].RemoveEntity(npc)
		g.awardGold(sourceEntityID)
		g.awardExperience(sourceEntityID)
	} else if skillID == d2hero.SkillEclatDeGlace {
		npc.ApplySlow(g.clock().Add(eclatDeGlaceSlowDuration))
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

// resolveNovaHit resolves Nova de givre: hits every killable NPC within
// novaDeGivreRadiusSubtiles of the caster's own position (read off their
// HeroState -- the server has no map-entity for players, see nearestPlayer).
// A no-op if sourceEntityID isn't a resolved connected player.
//
// Limite de test connue (same as resolveChainHit): exercising the actual
// multi-NPC AoE path needs live d2mapentity.NPC instances in a map engine,
// whose fields are private outside their own package -- only the no-op path
// is unit-tested here.
func (g *GameServer) resolveNovaHit(sourceEntityID string, skillID int) {
	state := g.playerStateOf(sourceEntityID)
	if state == nil {
		return
	}

	center := d2vector.NewPosition(state.X, state.Y)
	g.resolveAoeHit(center, sourceEntityID, skillID, novaDeGivreRadiusSubtiles)
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

	for _, entity := range g.mapEngines[0].Entities() {
		npc, ok := entity.(*d2mapentity.NPC)
		if !ok || !npc.IsKillable() || npc.HP <= 0 {
			continue
		}

		g.applyResolvedDamage(npc, sourceEntityID, skillID, npc.HP*champStatiqueDamagePercent/100)
	}
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
// ponytail: no elemental type, no gear modifiers (e.g. the "+10% dégâts
// Feu" a staff can grant), no attack rating. See ROADMAP.md Phase 1/2 for
// the rest of the combat formula.
func (g *GameServer) resolveAttackDamage(sourceEntityID string, skillID int) int {
	baseSort := d2hero.SkillBaseSortDamage(skillID, baseSortDamage)

	state := g.playerStateOf(sourceEntityID)
	if state == nil || state.Stats == nil {
		return baseSort
	}

	damage := baseSort + (baseSort*state.Stats.Energy)/100

	// ponytail: always applies the equipped weapon's Fire modifier,
	// regardless of the cast skill's actual element -- there's no per-skill
	// element data yet (d2hero.DevilSkillDef has no element field either).
	// See ROADMAP.md Phase 2/5.
	if firePercent := d2hero.ItemFireDamagePercent(state.Equipment.RightHand.GetItemCode()); firePercent > 0 {
		damage += (damage * firePercent) / 100
	}

	return damage
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
	clock               func() time.Time // overridden in tests; defaults to time.Now

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
