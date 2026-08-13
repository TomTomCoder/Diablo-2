package d2mapentity

import (
	"math/rand"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2path"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
)

// NPC is a passive complex entity with which the player can interact.
// For example, Deckard Cain.
type NPC struct {
	mapEntity
	Paths         []d2path.Path
	name          string
	composite     *d2asset.Composite
	action        int
	path          int
	repetitions   int
	monstatRecord *d2records.MonStatRecord
	monstatEx     *d2records.MonStat2Record
	HasPaths      bool
	isDone        bool
	HP            int

	// SlowedUntil is when this NPC's movement speed reduction expires (see
	// ApplySlow/IsSlowed). Zero value means not slowed.
	SlowedUntil time.Time

	// AmplifiedUntil is when this NPC's incoming-damage amplification
	// (Amplification, "Augmente les dégâts magiques reçus par la cible")
	// expires. Zero value means not amplified. See
	// ApplyAmplification/IsAmplified.
	AmplifiedUntil time.Time

	// ImmobilizedUntil is when this NPC's full immobilization (Prison de
	// glace, "Immobilise un groupe d'entités") expires. Zero value means
	// not immobilized. Unlike ApplySlow's partial speed reduction, an
	// immobilized NPC's ChasePlayer speed is forced to zero. See
	// ApplyImmobilize/IsImmobilized.
	ImmobilizedUntil time.Time

	// ResistanceStrippedUntil is when this NPC's magic resistance removal
	// (Rupture arcane, "Projectile qui supprime les résistances d'une
	// cible") expires. Zero value means resistance isn't stripped. See
	// ApplyResistanceStrip/IsResistanceStripped/MagicResistancePercent.
	ResistanceStrippedUntil time.Time
}

const (
	magicOffsetX            = 5
	magicOffsetScalarX      = 8
	magicOffsetScalarY      = 16
	minAnimationRepetitions = 3
	maxAnimationRepetitions = 5
)

func selectEquip(slice []string) string {
	if len(slice) != 0 {
		// nolint:gosec // not concerned with crypto-strong randomness
		return slice[rand.Intn(len(slice))]
	}

	return ""
}

// ID returns the NPC uuid
func (v *NPC) ID() string {
	return v.mapEntity.uuid
}

// IsKillable reports whether this NPC can take damage and die (matches
// monstats.txt's `killable` column -- town NPCs like Deckard Cain are not).
func (v *NPC) IsKillable() bool {
	return v.monstatRecord != nil && v.monstatRecord.IsKillable
}

// MonsterKey returns this NPC's stable monster-type identifier
// (d2records.MonStatRecord.Key, "Id" in monstats.txt), or "" if it has no
// monstat record. Meant as a lookup key into a per-monster-type data
// registry (e.g. d2hero.MonsterArchetypes) -- see
// GameServer.advanceMonsterAI/tryMonsterAttack.
func (v *NPC) MonsterKey() string {
	if v.monstatRecord == nil {
		return ""
	}

	return v.monstatRecord.Key
}

// MagicResistancePercent returns this NPC's magic resistance (monstats.txt's
// ResMa column, d2records.MonStatRecord.ResistanceMagicNormal), or 0 if it
// has no monstat record. Can be negative (takes more damage) or >=100
// (immune) -- real D2 semantics that d2hero.MitigateDamage already handles.
//
// Unlike the player's own resistances, this is NOT run through
// d2hero.CapResistance: that "never negative, capped at 75%" rule is stated
// in devil_game_design_reference.md §6 specifically for the Mage's own
// defense, not for the monsters Devil's world throws at them.
//
// Doesn't account for Rupture arcane's resistance removal itself -- callers
// combine this with IsResistanceStripped (mirroring how amplifiedDamage
// combines a base value with IsAmplified in game_server.go).
//
// ponytail: always Normal-difficulty, same limitation as
// MonStatRecord.HPRangeForDifficulty -- see that method's doc comment.
func (v *NPC) MagicResistancePercent() int {
	if v.monstatRecord == nil {
		return 0
	}

	return v.monstatRecord.ResistanceMagicNormal
}

// AttackDamageRange returns this NPC's primary melee attack's damage range
// (monstats.txt's A1MinD/A1MaxD columns, d2records.MonStatRecord.
// DamageMinA1Normal/DamageMaxA1Normal), or (0, 0) if it has no monstat
// record. Only the "A1" (first melee attack) columns are read -- Devil's
// own tryMonsterAttack deals a single flat amount per hit already, not
// modeling per-monster secondary attacks ("A2") or skill-based ones
// ("S1"), so reading more than one column here would have nowhere to go.
//
// ponytail: always Normal-difficulty, same limitation as
// MonStatRecord.HPRangeForDifficulty/NPC.MagicResistancePercent.
func (v *NPC) AttackDamageRange() (min, max int) {
	if v.monstatRecord == nil {
		return 0, 0
	}

	return v.monstatRecord.DamageMinA1Normal, v.monstatRecord.DamageMaxA1Normal
}

// IsColdImmune reports whether this NPC is immune to cold-elemental slow
// effects (monstats.txt's ColdSensitivityNormal column, "coldeffect" --
// d2records.MonStatRecord's own doc comment: "0 = ... unfreezeable"). false
// (not immune) if it has no monstat record.
//
// Narrowly scoped to Éclat de glace specifically (Devil's only
// Élémentalisme cold-damage skill that applies a slow -- Nova de
// givre/Orbe glaciale deal cold damage but never slow, and Ralentissement/
// Prison de glace/Distorsion temporelle are Arcane control effects, not
// cold-elemental damage, so a monster's cold sensitivity has no bearing on
// them). See ROADMAP.md for why the scope stops there.
//
// ponytail: always Normal-difficulty, same limitation as
// MonStatRecord.HPRangeForDifficulty/NPC.MagicResistancePercent.
func (v *NPC) IsColdImmune() bool {
	if v.monstatRecord == nil {
		return false
	}

	return v.monstatRecord.ColdSensitivityNormal == 0
}

// AggroDistanceTiles returns this NPC's AI activation radius in tiles
// (monstats.txt's "aidist" column, d2records.MonStatRecord.AiDistanceNormal),
// or 0 if it has no monstat record or the column was left blank. 0 is a
// safe "no data" sentinel: real D2 treats a blank aidist as an implicit ~35
// (that default lives in the game engine, not in the data file), and
// Devil's own caller already falls back to its own flat radius constant
// whenever this returns 0 -- same convention as AttackDamageRange.
//
// ponytail: always Normal-difficulty, same limitation as
// MonStatRecord.HPRangeForDifficulty/NPC.MagicResistancePercent.
func (v *NPC) AggroDistanceTiles() int {
	if v.monstatRecord == nil {
		return 0
	}

	return v.monstatRecord.AiDistanceNormal
}

// ApplyDamage reduces the NPC's HP by amount and reports whether it died.
// No-op (and never dies) for NPCs that aren't killable.
func (v *NPC) ApplyDamage(amount int) (died bool) {
	if !v.IsKillable() || v.HP <= 0 {
		return v.IsKillable() && v.HP <= 0
	}

	v.HP -= amount

	if v.HP < 0 {
		v.HP = 0
	}

	return v.HP == 0
}

// npcChaseSpeed is the movement speed a killable NPC uses while chasing a
// player. ponytail: one flat speed for every monster -- no per-monster
// speed data exists yet (ROADMAP.md Phase 4).
const npcChaseSpeed = 6.0

// ChasePlayer starts this NPC moving directly toward pos.
//
// ponytail: straight-line movement (setTarget), not real pathfinding --
// an NPC will walk straight through obstacles between it and the player
// rather than routing around them. Fine for open areas; see ROADMAP.md
// Phase 4 for routing around walls.
func (v *NPC) ChasePlayer(pos d2vector.Position, now time.Time) {
	if v.IsImmobilized(now) {
		v.SetSpeed(0)
		return
	}

	speed := npcChaseSpeed
	if v.IsSlowed(now) {
		speed *= npcSlowedSpeedMultiplier
	}

	v.SetSpeed(speed)
	v.setTarget(pos, nil)
}

// npcSlowedSpeedMultiplier is how much ChasePlayer's speed is reduced by
// while SlowedUntil hasn't elapsed (e.g. after being hit by Éclat de
// glace, "ralentit la cible" -- devil_game_design_reference.md §7).
const npcSlowedSpeedMultiplier = 0.5

// ApplySlow marks this NPC as slowed until the given time.
func (v *NPC) ApplySlow(until time.Time) {
	v.SlowedUntil = until
}

// IsSlowed reports whether this NPC's slow effect is still active at now.
func (v *NPC) IsSlowed(now time.Time) bool {
	return now.Before(v.SlowedUntil)
}

// ApplyAmplification marks this NPC as amplified (taking increased damage,
// see IsAmplified) until the given time.
func (v *NPC) ApplyAmplification(until time.Time) {
	v.AmplifiedUntil = until
}

// IsAmplified reports whether this NPC's damage amplification is still
// active at now.
func (v *NPC) IsAmplified(now time.Time) bool {
	return now.Before(v.AmplifiedUntil)
}

// ApplyImmobilize marks this NPC as fully immobilized (see IsImmobilized)
// until the given time.
func (v *NPC) ApplyImmobilize(until time.Time) {
	v.ImmobilizedUntil = until
}

// ApplyResistanceStrip marks this NPC's magic resistance as removed (see
// IsResistanceStripped/MagicResistancePercent) until the given time.
func (v *NPC) ApplyResistanceStrip(until time.Time) {
	v.ResistanceStrippedUntil = until
}

// IsResistanceStripped reports whether this NPC's magic resistance removal
// is still active at now.
func (v *NPC) IsResistanceStripped(now time.Time) bool {
	return now.Before(v.ResistanceStrippedUntil)
}

// IsImmobilized reports whether this NPC's immobilization is still active
// at now.
func (v *NPC) IsImmobilized(now time.Time) bool {
	return now.Before(v.ImmobilizedUntil)
}

// Knockback instantly displaces this NPC distanceSubtiles further away from
// source, along the line from source through its current position. Falls
// back to pushing along +X if source and the NPC's position coincide
// (nothing to normalize a direction from).
//
// ponytail: an instant teleport-away, not real push-back physics -- no
// animation, and no collision check against walls or other entities. Good
// enough for Télékinésie's design ("repousse les entités") until real
// knockback exists.
func (v *NPC) Knockback(source d2vector.Position, distanceSubtiles float64) {
	direction := v.Position.Clone().Subtract(&source.Vector)
	if direction.IsZero() {
		direction = d2vector.VectorRight()
	}

	direction.SetLength(distanceSubtiles)
	v.Position.Set(v.Position.X()+direction.X(), v.Position.Y()+direction.Y())
}

// Pull instantly displaces this NPC distanceSubtiles closer to target, along
// the line from its current position to target -- Knockback's inverse.
// Clamped so it never overshoots past target itself. A no-op if the NPC is
// already at target (nothing to normalize a direction from).
//
// ponytail: an instant teleport-closer, not real pull physics -- no
// animation, and no collision check against walls or other entities. Good
// enough for Vortex's design ("aspire ... vers un point") until real pull
// physics exist.
func (v *NPC) Pull(target d2vector.Position, distanceSubtiles float64) {
	direction := target.Clone().Subtract(&v.Position.Vector)

	dist := direction.Length()
	if dist == 0 {
		return
	}

	if distanceSubtiles > dist {
		distanceSubtiles = dist
	}

	direction.SetLength(distanceSubtiles)
	v.Position.Set(v.Position.X()+direction.X(), v.Position.Y()+direction.Y())
}

// Render renders this entity's animated composite.
func (v *NPC) Render(target d2interface.Surface) {
	renderOffset := v.Position.RenderOffset()
	target.PushTranslation(
		int((renderOffset.X()-renderOffset.Y())*magicOffsetScalarY),
		int(((renderOffset.X()+renderOffset.Y())*magicOffsetScalarX)-magicOffsetX),
	)

	defer target.Pop()

	if v.composite.Render(target) != nil {
		return
	}
}

// Path returns the current part of the entity's path.
func (v *NPC) Path() d2path.Path {
	return v.Paths[v.path]
}

// NextPath returns the next part of the entity's path.
func (v *NPC) NextPath() d2path.Path {
	v.path++
	if v.path == len(v.Paths) {
		v.path = 0
	}

	return v.Paths[v.path]
}

// SetPaths sets the entity's paths to the given slice. It also sets flags
// on the entity indicating that it has paths and has completed the
// previous none.
func (v *NPC) SetPaths(paths []d2path.Path) {
	v.Paths = paths
	v.HasPaths = len(paths) > 0
	v.isDone = true
}

// Advance is called once per frame and processes a
// single game tick.
func (v *NPC) Advance(tickTime float64) {
	v.Step(tickTime)

	if err := v.composite.Advance(tickTime); err != nil {
		return
	}

	if v.HasPaths && v.wait() {
		// If at the target, set target to the next path.
		v.isDone = false
		path := v.NextPath()
		v.setTarget(
			path.Position,
			v.next,
		)

		v.action = path.Action
	}
}

// If an npc has a path to pause at each location.
// Waits for animation to end and all repetitions to be exhausted.
func (v *NPC) wait() bool {
	return v.isDone && v.composite.GetPlayedCount() > v.repetitions
}

func (v *NPC) next() {
	var newAnimationMode d2enum.MonsterAnimationMode

	v.isDone = true

	// nolint:gosec // not concerned with crypto-strong randomness
	v.repetitions = minAnimationRepetitions + rand.Intn(maxAnimationRepetitions)

	switch d2enum.NPCActionType(v.action) {
	case d2enum.NPCActionSkill1:
		newAnimationMode = d2enum.MonsterAnimationModeSkill1
		v.repetitions = 0
	case d2enum.NPCActionInvalid, d2enum.NPCAction1, d2enum.NPCAction2, d2enum.NPCAction3:
		newAnimationMode = d2enum.MonsterAnimationModeNeutral
		v.repetitions = 0
	default:
		newAnimationMode = d2enum.MonsterAnimationModeNeutral
		v.repetitions = 0
	}

	if v.composite.GetAnimationMode() != newAnimationMode.String() {
		if err := v.composite.SetMode(newAnimationMode, v.composite.GetWeaponClass()); err != nil {
			return
		}
	}
}

// rotate sets direction and changes animation
func (v *NPC) rotate(direction int) {
	var newMode d2enum.MonsterAnimationMode
	if !v.atTarget() {
		newMode = d2enum.MonsterAnimationModeWalk
	} else {
		newMode = d2enum.MonsterAnimationModeNeutral
	}

	if newMode.String() != v.composite.GetAnimationMode() {
		if err := v.composite.SetMode(newMode, v.composite.GetWeaponClass()); err != nil {
			return
		}
	}

	if v.composite.GetDirection() != direction {
		v.composite.SetDirection(direction)
	}
}

// Selectable returns true if the object can be highlighted/selected.
func (v *NPC) Selectable() bool {
	// is there something handy that determines selectable npc's?
	return v.name != ""
}

// Label returns the NPC's in-game name (e.g. "Deckard Cain") or an empty string if it does not have a name.
func (v *NPC) Label() string {
	return v.name
}

// GetPosition returns the NPC's position
func (v *NPC) GetPosition() d2vector.Position {
	return v.mapEntity.Position
}

// GetVelocity returns the NPC's velocity vector
func (v *NPC) GetVelocity() d2vector.Vector {
	return v.mapEntity.velocity
}

// GetSize returns the current frame size
func (v *NPC) GetSize() (width, height int) {
	return v.composite.GetSize()
}
