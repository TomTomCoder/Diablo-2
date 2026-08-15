package d2mapengine

import (
	"sync"
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2geom"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
)

// fakeEntity is a minimal d2interface.MapEntity for tests that don't need a
// real asset manager -- these tests exercise MapEngine.entities' own
// bookkeeping (Entities/AddEntity/RemoveEntity/Advance), not rendering or
// real game data, so a real *d2mapentity.NPC/Missile (which needs an
// AssetManager loaded from real MPQ files, unavailable in this
// environment) isn't needed.
type fakeEntity struct {
	id        string
	onAdvance func()
}

func (f *fakeEntity) ID() string                       { return f.id }
func (f *fakeEntity) Render(_ d2interface.Surface)     {}
func (f *fakeEntity) GetPosition() d2vector.Position   { return d2vector.Position{} }
func (f *fakeEntity) GetVelocity() d2vector.Vector     { return d2vector.Vector{} }
func (f *fakeEntity) GetSize() (width, height int)     { return 0, 0 }
func (f *fakeEntity) GetLayer() int                    { return 0 }
func (f *fakeEntity) GetPositionF() (float64, float64) { return 0, 0 }
func (f *fakeEntity) Label() string                    { return "" }
func (f *fakeEntity) Selectable() bool                 { return false }
func (f *fakeEntity) Highlight()                       {}
func (f *fakeEntity) Advance(tickTime float64) {
	if f.onAdvance != nil {
		f.onAdvance()
	}
}

func TestMarkExploredAndIsExplored(t *testing.T) {
	m := &MapEngine{size: d2geom.Size{Width: 10, Height: 10}, explored: make([]bool, 100)}

	if m.IsExplored(3, 4) {
		t.Fatal("expected a fresh map to have no explored tiles")
	}

	m.MarkExplored(3, 4)

	if !m.IsExplored(3, 4) {
		t.Error("expected the marked tile to report explored")
	}

	if m.IsExplored(3, 5) {
		t.Error("expected a neighboring, never-marked tile to still report unexplored")
	}
}

// TestMarkExploredOutOfBoundsIsNoop and TestIsExploredOutOfBoundsIsFalse
// mirror TileAt's own out-of-bounds handling (nil rather than a panic) --
// MarkExplored/IsExplored must never index-panic on a coordinate outside
// the map, e.g. from a player standing at the very edge of a level.
func TestMarkExploredOutOfBoundsIsNoop(t *testing.T) {
	m := &MapEngine{size: d2geom.Size{Width: 10, Height: 10}, explored: make([]bool, 100)}

	// must not panic.
	m.MarkExplored(-1, -1)
	m.MarkExplored(100, 100)
}

func TestIsExploredOutOfBoundsIsFalse(t *testing.T) {
	m := &MapEngine{size: d2geom.Size{Width: 10, Height: 10}, explored: make([]bool, 100)}

	if m.IsExplored(-1, -1) || m.IsExplored(100, 100) {
		t.Error("expected out-of-bounds coordinates to report unexplored, not panic or report true")
	}
}

// TestEntitiesReturnsASnapshotNotTheLiveMap is a regression test for the
// concurrency-safety correction: Entities() used to return the live
// m.entities map itself, so a caller ranging it while another goroutine
// called RemoveEntity could crash the process
// ("fatal error: concurrent map iteration and map write"). Verifies the
// returned map is an independent copy: mutating MapEngine's real entities
// afterward must not change what was already returned.
func TestEntitiesReturnsASnapshotNotTheLiveMap(t *testing.T) {
	m := &MapEngine{entities: make(map[string]d2interface.MapEntity)}
	m.AddEntity(&fakeEntity{id: "a"})

	snapshot := m.Entities()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 entity in the snapshot, got %d", len(snapshot))
	}

	m.RemoveEntity(&fakeEntity{id: "a"})

	if len(snapshot) != 1 {
		t.Errorf("expected the earlier snapshot to be unaffected by a later RemoveEntity, got %d entries", len(snapshot))
	}

	if len(m.Entities()) != 0 {
		t.Errorf("expected a fresh Entities() call to reflect the removal, got %d entries", len(m.Entities()))
	}
}

// TestAdvanceHandlesSelfRemovingEntity is a regression test for the
// reentrancy hazard the Advance fix specifically avoids: a real Missile's
// Advance can synchronously call RemoveEntity on itself (via
// mapEntity.Step's "if the path is complete it calls entity.done()").
// Advance releases entitiesMu before calling any entity's own Advance for
// exactly this reason -- holding it would make this a reentrant Lock() on
// a non-reentrant sync.Mutex, an unconditional deadlock. This test doesn't
// prove deadlock absence on its own (a hung goroutine wouldn't fail the
// test directly), but go test's own default timeout does: a deadlock here
// would time out the whole test binary rather than silently pass.
func TestAdvanceHandlesSelfRemovingEntity(t *testing.T) {
	m := &MapEngine{entities: make(map[string]d2interface.MapEntity)}

	self := &fakeEntity{id: "missile"}
	self.onAdvance = func() { m.RemoveEntity(self) }

	m.AddEntity(self)
	m.Advance(1.0 / 60)

	if len(m.Entities()) != 0 {
		t.Errorf("expected the self-removing entity to be gone after Advance, got %d entries", len(m.Entities()))
	}
}

// TestEntitiesRaceWithAddRemoveEntity is a regression test exercised under
// `go test -race`: one goroutine repeatedly calls Entities() (mirroring
// GameServer.advanceMonsterAI's AI-tick ticker) while another repeatedly
// calls AddEntity/RemoveEntity (mirroring the packet-dispatch goroutine's
// resolve* handlers) on the same MapEngine. Before the correction, this
// would reliably trip the race detector (or, outside -race, risk the
// "fatal error: concurrent map iteration and map write" crash the finding
// described) -- entitiesMu now serializes every access.
func TestEntitiesRaceWithAddRemoveEntity(t *testing.T) {
	m := &MapEngine{entities: make(map[string]d2interface.MapEntity)}

	const iterations = 500

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < iterations; i++ {
			for _, entity := range m.Entities() {
				_ = entity.ID()
			}
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < iterations; i++ {
			e := &fakeEntity{id: "e"}
			m.AddEntity(e)
			m.RemoveEntity(e)
		}
	}()

	wg.Wait()
}
