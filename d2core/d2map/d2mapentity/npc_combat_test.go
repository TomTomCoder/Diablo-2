package d2mapentity

import (
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
