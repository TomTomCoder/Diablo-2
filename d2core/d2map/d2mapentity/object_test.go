package d2mapentity

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

func TestObjectIsWaypoint(t *testing.T) {
	waypoint := &Object{objectRecord: &d2records.ObjectDetailRecord{InitFn: objectInitFnWaypoint}}

	if !waypoint.IsWaypoint() {
		t.Error("expected an object with the waypoint InitFn to report IsWaypoint true")
	}

	torch := &Object{objectRecord: &d2records.ObjectDetailRecord{InitFn: 8}}

	if torch.IsWaypoint() {
		t.Error("expected a non-waypoint object (e.g. a torch) to report IsWaypoint false")
	}
}

func TestObjectIsWaypointNilRecordIsFalse(t *testing.T) {
	ob := &Object{}

	if ob.IsWaypoint() {
		t.Error("expected an object with no objectRecord to report IsWaypoint false, not panic")
	}
}

func TestWaypointTilesFindsOnlyWaypoints(t *testing.T) {
	waypoint := &Object{
		objectRecord: &d2records.ObjectDetailRecord{InitFn: objectInitFnWaypoint},
		Position:     d2vector.NewPositionTile(5, 6),
	}
	torch := &Object{
		objectRecord: &d2records.ObjectDetailRecord{InitFn: 8},
		Position:     d2vector.NewPositionTile(1, 1),
	}

	entities := map[string]d2interface.MapEntity{
		"waypoint": waypoint,
		"torch":    torch,
	}

	tiles := WaypointTiles(entities)

	if len(tiles) != 1 {
		t.Fatalf("expected exactly 1 waypoint tile (the torch excluded), got %d: %v", len(tiles), tiles)
	}

	if tiles[0] != [2]int{5, 6} {
		t.Errorf("expected the waypoint's own tile (5,6), got %v", tiles[0])
	}
}

func TestWaypointTilesNoWaypointsIsEmpty(t *testing.T) {
	torch := &Object{objectRecord: &d2records.ObjectDetailRecord{InitFn: 8}}

	tiles := WaypointTiles(map[string]d2interface.MapEntity{"torch": torch})

	if len(tiles) != 0 {
		t.Errorf("expected no waypoint tiles among non-waypoint entities, got %v", tiles)
	}
}
