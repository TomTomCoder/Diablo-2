package d2mapentity

import (
	"math"
	"testing"
	"time"

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
