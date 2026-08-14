package d2mapentity

import (
	"math"
	"testing"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

func killableNPC(hp int) *NPC {
	return &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{IsKillable: true},
		HP:            hp,
	}
}

func TestNPCIsKillable(t *testing.T) {
	killable := killableNPC(10)
	if !killable.IsKillable() {
		t.Error("expected NPC with IsKillable monstat to be killable")
	}

	townNPC := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{IsKillable: false},
		HP:            10,
	}
	if townNPC.IsKillable() {
		t.Error("expected NPC with IsKillable=false monstat to not be killable")
	}
}

func TestNPCMonsterKey(t *testing.T) {
	npc := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{IsKillable: true, Key: "fallen1"},
		HP:            10,
	}

	if got, want := npc.MonsterKey(), "fallen1"; got != want {
		t.Errorf("expected MonsterKey %q, got %q", want, got)
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	if got := noMonstat.MonsterKey(); got != "" {
		t.Errorf("expected empty MonsterKey with no monstat record, got %q", got)
	}
}

func TestNPCMagicResistancePercent(t *testing.T) {
	resistant := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{ResistanceMagicNormal: 50},
	}
	if got, want := resistant.MagicResistancePercent(d2enum.DifficultyNormal), 50; got != want {
		t.Errorf("expected MagicResistancePercent %d, got %d", want, got)
	}

	vulnerable := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{ResistanceMagicNormal: -50},
	}
	if got, want := vulnerable.MagicResistancePercent(d2enum.DifficultyNormal), -50; got != want {
		t.Errorf("expected negative MagicResistancePercent %d (takes more damage), got %d", want, got)
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	if got := noMonstat.MagicResistancePercent(d2enum.DifficultyNormal); got != 0 {
		t.Errorf("expected MagicResistancePercent 0 with no monstat record, got %d", got)
	}
}

// TestNPCMagicResistancePercentVariesByDifficulty is a regression test:
// MagicResistancePercent used to always read ResistanceMagicNormal
// regardless of the difficulty passed in (devil_game_design_reference.md
// §3: "leurs résistances augmentent modérément" at Nightmare/Hell).
func TestNPCMagicResistancePercentVariesByDifficulty(t *testing.T) {
	npc := &NPC{
		mapEntity: newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{
			ResistanceMagicNormal:    10,
			ResistanceMagicNightmare: 30,
			ResistanceMagicHell:      50,
		},
	}

	if got, want := npc.MagicResistancePercent(d2enum.DifficultyNightmare), 30; got != want {
		t.Errorf("expected Nightmare resistance %d, got %d", want, got)
	}

	if got, want := npc.MagicResistancePercent(d2enum.DifficultyHell), 50; got != want {
		t.Errorf("expected Hell resistance %d, got %d", want, got)
	}
}

func TestNPCAttackDamageRange(t *testing.T) {
	npc := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{DamageMinA1Normal: 3, DamageMaxA1Normal: 7},
	}

	gotMin, gotMax := npc.AttackDamageRange(d2enum.DifficultyNormal)
	if gotMin != 3 || gotMax != 7 {
		t.Errorf("expected AttackDamageRange (3, 7), got (%d, %d)", gotMin, gotMax)
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	gotMin, gotMax = noMonstat.AttackDamageRange(d2enum.DifficultyNormal)
	if gotMin != 0 || gotMax != 0 {
		t.Errorf("expected AttackDamageRange (0, 0) with no monstat record, got (%d, %d)", gotMin, gotMax)
	}
}

// TestNPCAttackDamageRangeVariesByDifficulty mirrors
// TestNPCMagicResistancePercentVariesByDifficulty for the damage range.
func TestNPCAttackDamageRangeVariesByDifficulty(t *testing.T) {
	npc := &NPC{
		mapEntity: newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{
			DamageMinA1Normal: 3, DamageMaxA1Normal: 7,
			DamageMinA1Hell: 20, DamageMaxA1Hell: 40,
		},
	}

	gotMin, gotMax := npc.AttackDamageRange(d2enum.DifficultyHell)
	if gotMin != 20 || gotMax != 40 {
		t.Errorf("expected Hell AttackDamageRange (20, 40), got (%d, %d)", gotMin, gotMax)
	}
}

func TestNPCIsColdImmune(t *testing.T) {
	immune := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{ColdSensitivityNormal: 0},
	}
	if !immune.IsColdImmune() {
		t.Error("expected ColdSensitivityNormal=0 to mean cold immune")
	}

	vulnerable := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{ColdSensitivityNormal: 100},
	}
	if vulnerable.IsColdImmune() {
		t.Error("expected a nonzero ColdSensitivityNormal to not be immune")
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	if noMonstat.IsColdImmune() {
		t.Error("expected IsColdImmune false with no monstat record")
	}
}

func TestNPCAggroDistanceTiles(t *testing.T) {
	npc := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{AiDistanceNormal: 12},
	}
	if got := npc.AggroDistanceTiles(); got != 12 {
		t.Errorf("expected AggroDistanceTiles 12, got %d", got)
	}

	blank := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{},
	}
	if got := blank.AggroDistanceTiles(); got != 0 {
		t.Errorf("expected AggroDistanceTiles 0 for a blank aidist column, got %d", got)
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	if got := noMonstat.AggroDistanceTiles(); got != 0 {
		t.Errorf("expected AggroDistanceTiles 0 with no monstat record, got %d", got)
	}
}

func TestNPCAiDelayFrames(t *testing.T) {
	npc := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{AiDelayNormal: 9},
	}
	if got := npc.AiDelayFrames(); got != 9 {
		t.Errorf("expected AiDelayFrames 9, got %d", got)
	}

	blank := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{},
	}
	if got := blank.AiDelayFrames(); got != 0 {
		t.Errorf("expected AiDelayFrames 0 for a blank aidel column, got %d", got)
	}

	noMonstat := &NPC{mapEntity: newMapEntity(0, 0)}
	if got := noMonstat.AiDelayFrames(); got != 0 {
		t.Errorf("expected AiDelayFrames 0 with no monstat record, got %d", got)
	}
}

func TestNPCApplyDamage(t *testing.T) {
	npc := killableNPC(10)

	if died := npc.ApplyDamage(4); died {
		t.Error("NPC with 10 HP should not die from 4 damage")
	}

	if npc.HP != 6 {
		t.Errorf("expected HP 6 after 4 damage, got %d", npc.HP)
	}

	if died := npc.ApplyDamage(100); !died {
		t.Error("NPC should die when damage exceeds remaining HP")
	}

	if npc.HP != 0 {
		t.Errorf("expected HP to floor at 0, got %d", npc.HP)
	}

	// already dead: further damage is a no-op, and it's still reported as dead
	if died := npc.ApplyDamage(1); !died {
		t.Error("applying damage to an already-dead NPC should still report died=true")
	}
}

func TestNPCApplySlow(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()

	if npc.IsSlowed(now) {
		t.Fatal("a fresh NPC should not start slowed")
	}

	npc.ApplySlow(now.Add(time.Second))

	if !npc.IsSlowed(now) {
		t.Error("expected the NPC to be slowed immediately after ApplySlow")
	}

	if npc.IsSlowed(now.Add(2 * time.Second)) {
		t.Error("expected the slow to have expired after its duration elapsed")
	}
}

// TestNPCApplySlowNeverShortensAnAlreadyLongerWindow is a regression test:
// ApplySlow is shared by 4 casts with different flat durations (Éclat de
// glace/Ralentissement/Armure de glace's counter-slow all 3s, Distorsion
// temporelle 5s) -- it used to unconditionally overwrite SlowedUntil, so
// casting Distorsion temporelle then immediately Éclat de glace on the same
// target cut its slow down to 3s instead of leaving the longer ~5s window
// intact.
func TestNPCApplySlowNeverShortensAnAlreadyLongerWindow(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()

	npc.ApplySlow(now.Add(5 * time.Second)) // e.g. Distorsion temporelle
	npc.ApplySlow(now.Add(3 * time.Second)) // e.g. Éclat de glace, shortly after

	if !npc.IsSlowed(now.Add(4 * time.Second)) {
		t.Error("expected the longer 5s window to survive a shorter, later ApplySlow call")
	}
}

func TestNPCChasePlayerRespectsSlow(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()
	dest := d2vector.NewPosition(5, 5)

	npc.ChasePlayer(dest, now)
	normalSpeed := npc.GetSpeed()

	npc.ApplySlow(now.Add(time.Second))
	npc.ChasePlayer(dest, now)
	slowedSpeed := npc.GetSpeed()

	if slowedSpeed >= normalSpeed {
		t.Errorf("expected a slowed chase speed (%v) to be less than normal (%v)", slowedSpeed, normalSpeed)
	}

	if got, want := slowedSpeed, normalSpeed*npcSlowedSpeedMultiplier; got != want {
		t.Errorf("expected slowed speed %v, got %v", want, got)
	}
}

func TestNPCApplyImmobilize(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()

	if npc.IsImmobilized(now) {
		t.Fatal("a fresh NPC should not start immobilized")
	}

	npc.ApplyImmobilize(now.Add(time.Second))

	if !npc.IsImmobilized(now) {
		t.Error("expected the NPC to be immobilized immediately after ApplyImmobilize")
	}

	if npc.IsImmobilized(now.Add(2 * time.Second)) {
		t.Error("expected the immobilization to have expired after its duration elapsed")
	}
}

func TestNPCChasePlayerRespectsImmobilize(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()
	dest := d2vector.NewPosition(5, 5)

	npc.ApplyImmobilize(now.Add(time.Second))
	npc.ChasePlayer(dest, now)

	if got := npc.GetSpeed(); got != 0 {
		t.Errorf("expected an immobilized NPC to have speed 0, got %v", got)
	}
}

func TestNPCChasePlayerImmobilizeOverridesSlow(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()
	dest := d2vector.NewPosition(5, 5)

	// both effects active: immobilize should win outright (speed 0), not
	// just apply the slow multiplier on top.
	npc.ApplySlow(now.Add(time.Second))
	npc.ApplyImmobilize(now.Add(time.Second))
	npc.ChasePlayer(dest, now)

	if got := npc.GetSpeed(); got != 0 {
		t.Errorf("expected immobilize to override slow and force speed 0, got %v", got)
	}
}

func TestNPCApplyAmplification(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()

	if npc.IsAmplified(now) {
		t.Fatal("a fresh NPC should not start amplified")
	}

	npc.ApplyAmplification(now.Add(time.Second))

	if !npc.IsAmplified(now) {
		t.Error("expected the NPC to be amplified immediately after ApplyAmplification")
	}

	if npc.IsAmplified(now.Add(2 * time.Second)) {
		t.Error("expected the amplification to have expired after its duration elapsed")
	}
}

func TestNPCApplyResistanceStrip(t *testing.T) {
	npc := killableNPC(10)
	now := time.Now()

	if npc.IsResistanceStripped(now) {
		t.Fatal("a fresh NPC should not start resistance-stripped")
	}

	npc.ApplyResistanceStrip(now.Add(time.Second))

	if !npc.IsResistanceStripped(now) {
		t.Error("expected the NPC to be resistance-stripped immediately after ApplyResistanceStrip")
	}

	if npc.IsResistanceStripped(now.Add(2 * time.Second)) {
		t.Error("expected the resistance strip to have expired after its duration elapsed")
	}
}

func TestNPCKnockbackPushesAwayAlongSourceLine(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(5, 5)
	source := d2vector.NewPosition(0, 5) // due "left" of the NPC

	npc.Knockback(source, 3)

	// pushed further along the same +X line, distance from source unchanged
	// in Y, up by exactly 3 in X.
	if got, want := npc.Position.X(), 8.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected X=%v after knockback, got %v", want, got)
	}

	if got, want := npc.Position.Y(), 5.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected Y unchanged at %v, got %v", want, got)
	}
}

func TestNPCKnockbackIncreasesDistanceFromSource(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(5, 5)
	source := d2vector.NewPosition(2, 3)

	before := npc.Position.Distance(&source.Vector)
	npc.Knockback(source, 3)
	after := npc.Position.Distance(&source.Vector)

	if after <= before {
		t.Errorf("expected knockback to increase distance from source, before=%v after=%v", before, after)
	}
}

func TestNPCKnockbackFromSamePositionDoesNotPanic(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(5, 5)

	// source coincides with the NPC's own position -- no direction to
	// normalize; must fall back rather than panic (e.g. on SetLength of a
	// zero-length vector).
	npc.Knockback(npc.Position, 3)
}

func TestNPCPullMovesTowardTarget(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(10, 5)
	target := d2vector.NewPosition(0, 5)

	npc.Pull(target, 3)

	// pulled 3 subtiles closer along the same -X line toward target
	if got, want := npc.Position.X(), 7.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected X=%v after pull, got %v", want, got)
	}

	if got, want := npc.Position.Y(), 5.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected Y unchanged at %v, got %v", want, got)
	}
}

func TestNPCPullClampsAtTarget(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(2, 5)
	target := d2vector.NewPosition(0, 5)

	npc.Pull(target, 10) // far more than the actual distance (2)

	if got, want := npc.Position.X(), 0.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected pull to clamp exactly at target X=%v, got %v", want, got)
	}

	if got, want := npc.Position.Y(), 5.0; math.Abs(got-want) > 0.001 {
		t.Errorf("expected Y unchanged at %v, got %v", want, got)
	}
}

func TestNPCPullAtTargetIsNoop(t *testing.T) {
	npc := killableNPC(10)
	npc.Position = d2vector.NewPosition(5, 5)

	// already at target -- no direction to normalize; must no-op rather
	// than panic.
	npc.Pull(npc.Position, 3)

	if npc.Position.X() != 5 || npc.Position.Y() != 5 {
		t.Errorf("expected position unchanged, got (%v, %v)", npc.Position.X(), npc.Position.Y())
	}
}

func TestNPCApplyDamageNotKillable(t *testing.T) {
	townNPC := &NPC{
		mapEntity:     newMapEntity(0, 0),
		monstatRecord: &d2records.MonStatRecord{IsKillable: false},
		HP:            10,
	}

	if died := townNPC.ApplyDamage(1000); died {
		t.Error("a non-killable NPC must never die, regardless of damage")
	}

	if townNPC.HP != 10 {
		t.Errorf("a non-killable NPC's HP should be untouched, got %d", townNPC.HP)
	}
}
