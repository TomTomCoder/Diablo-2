package d2player

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapengine"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
)

func TestMiniMapToggle(t *testing.T) {
	m := &MiniMap{}

	if m.Visible {
		t.Fatal("expected a fresh MiniMap to start hidden")
	}

	m.Toggle()

	if !m.Visible {
		t.Error("expected the first toggle to make it visible")
	}

	m.Toggle()

	if m.Visible {
		t.Error("expected the second toggle to hide it again")
	}
}

func TestMiniMapWindowClampsToMapBounds(t *testing.T) {
	// Player well inside the map: the window isn't clamped at all.
	minX, minY, maxX, maxY := miniMapWindow(50, 50, 100, 100)
	if minX != 30 || minY != 30 || maxX != 70 || maxY != 70 {
		t.Errorf("expected an unclamped 20-tile radius window, got (%d,%d)-(%d,%d)", minX, minY, maxX, maxY)
	}

	// Player near the top-left corner: clamped to 0 on both axes.
	minX, minY, _, _ = miniMapWindow(5, 5, 100, 100)
	if minX != 0 || minY != 0 {
		t.Errorf("expected the window clamped to 0 near the top-left corner, got min=(%d,%d)", minX, minY)
	}

	// Player near the bottom-right corner: clamped to the map's own size.
	_, _, maxX, maxY = miniMapWindow(98, 98, 100, 100)
	if maxX != 100 || maxY != 100 {
		t.Errorf("expected the window clamped to the map size near the bottom-right corner, got max=(%d,%d)", maxX, maxY)
	}
}

func TestMiniMapAdvanceNilInputsIsNoop(t *testing.T) {
	m := &MiniMap{}

	// must not panic with no local player/map engine resolved yet.
	m.Advance(nil, nil)
	m.Advance(&d2mapengine.MapEngine{}, nil)
	m.Advance(nil, &d2mapentity.Player{})
}

func TestMiniMapAdvanceMarksTilesAroundPlayerExplored(t *testing.T) {
	engine := d2mapengine.NewSizedMapEngine(20, 20)
	player := &d2mapentity.Player{}
	player.Position = d2vector.NewPositionTile(10, 10)

	m := &MiniMap{}
	m.Advance(engine, player)

	if !engine.IsExplored(10, 10) {
		t.Error("expected the player's own tile to be marked explored")
	}

	if !engine.IsExplored(10+miniMapExploreRadiusTiles, 10) {
		t.Error("expected a tile at the edge of the explore radius to be marked explored")
	}

	if engine.IsExplored(10+miniMapExploreRadiusTiles+1, 10) {
		t.Error("expected a tile just outside the explore radius to stay unexplored")
	}
}

func TestMiniMapRenderNilInputsIsNoop(t *testing.T) {
	surface := &fakeSurface{}
	m := &MiniMap{Visible: true}

	// must not panic with no local player/map engine resolved yet.
	m.Render(surface, nil, nil)
	m.Render(surface, &d2mapengine.MapEngine{}, nil)
	m.Render(surface, nil, &d2mapentity.Player{})

	if len(surface.rects) != 0 {
		t.Errorf("expected no draw calls with a nil input, got %v", surface.rects)
	}
}

func TestMiniMapRenderNotVisibleIsNoop(t *testing.T) {
	surface := &fakeSurface{}
	engine := d2mapengine.NewSizedMapEngine(20, 20)
	player := &d2mapentity.Player{}
	player.Position = d2vector.NewPositionTile(10, 10)

	m := &MiniMap{Visible: false}
	m.Advance(engine, player)
	m.Render(surface, engine, player)

	if len(surface.rects) != 0 {
		t.Errorf("expected no draw calls while hidden, got %v", surface.rects)
	}
}

func TestMiniMapRenderDrawsExploredTilesAndPlayer(t *testing.T) {
	surface := &fakeSurface{}
	engine := d2mapengine.NewSizedMapEngine(20, 20)
	player := &d2mapentity.Player{}
	player.Position = d2vector.NewPositionTile(10, 10)

	m := &MiniMap{Visible: true}
	m.Advance(engine, player) // marks the player's own tile + radius explored
	m.Render(surface, engine, player)

	// background rect + at least one explored-tile dot (the player's own tile).
	if len(surface.rects) < 2 {
		t.Fatalf("expected at least a background rect and one tile dot, got %d DrawRect calls", len(surface.rects))
	}

	for _, width := range surface.rects[1:] {
		if width != miniMapTileScale {
			t.Errorf("expected every tile dot to be %d px wide, got %d", miniMapTileScale, width)
		}
	}
}
