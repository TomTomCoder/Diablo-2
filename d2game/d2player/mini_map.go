package d2player

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapengine"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
)

// MiniMap is a real, if minimal, automap overlay -- unblocking ROADMAP.md's
// "génération procédurale améliorée" §4.1 item, whose own d2enum.ToggleMiniMap/
// ToggleAutomap keybindings were declared (key_map.go, key_binding_menu.go)
// but never wired to anything, and whose mini-panel button
// (miniPanelActions.automapToggle) was left nil -- both dead ends pointing
// at a feature that was never actually built, not merely "incomplete" the
// way the design text assumed.
//
// Same practice as DevilHUD (its own doc comment): no real Diablo 2
// automap/minimap art exists in this environment (ROADMAP.md Phase 6), so
// this renders with Surface's own primitives (DrawRect) instead of
// inventing sprite art -- a flat top-down grid of dots, not an isometric
// projection (d2maprenderer's own WorldToOrtho/OrthoToScreen are deliberately
// not reused here, see miniMapWindow's own doc comment).
//
// ponytail: shows only a fixed-radius window around the player (
// miniMapWindowRadiusTiles), not the whole level scaled to fit -- correct
// for "always-visible corner overlay", not full-screen automap parity with
// stock D2 (Tab's traditional zoomed-out view). Extend if that's ever
// actually needed.
type MiniMap struct {
	// Visible is whether the overlay renders at all -- toggled by
	// ToggleMiniMap/ToggleAutomap (GameControls.OnKeyDown) and the
	// mini-panel's own automap button, both wired to the same Toggle
	// here rather than two separate features: the design never
	// describes "minimap" and "automap" as distinct mechanics, and
	// inventing two different behaviors for what real Diablo 2 treats
	// as one concept would be exactly the kind of unspecified-behavior
	// invention this session avoids.
	Visible bool
}

const (
	miniMapWindowRadiusTiles  = 20 // how many tiles around the player are drawn
	miniMapExploreRadiusTiles = 3  // how many tiles around the player get marked explored per Advance
	miniMapTileScale          = 3  // pixels per tile
	miniMapMarginX            = 10
	miniMapMarginY            = 10

	miniMapBackgroundColor = 0x00000090 // translucent black
	miniMapExploredColor   = 0x808080ff // opaque gray
	miniMapPlayerColor     = 0xffff00ff // opaque yellow
	miniMapWaypointColor   = 0x00c8ffff // opaque cyan, distinct from both of the above

	miniMapCenterSubtile = 2                    // GetSubTileFlags' own 5x5 grid center, see its doc comment
	miniMapWaypointScale = miniMapTileScale * 2 // larger than a plain tile dot, to stand out as a point of interest
)

// Toggle negates Visible.
func (m *MiniMap) Toggle() {
	m.Visible = !m.Visible
}

// Advance marks the tiles around player's current position as explored
// (see d2mapengine.MapEngine.MarkExplored) -- called once per game tick
// regardless of Visible, so toggling the overlay on later still shows
// everywhere already walked through, matching real Diablo 2's own
// automap behavior of remembering exploration while hidden. A no-op if
// mapEngine or player is nil (no local player/map resolved yet, same
// guard as DevilHUD's own Render).
func (m *MiniMap) Advance(mapEngine *d2mapengine.MapEngine, player *d2mapentity.Player) {
	if mapEngine == nil || player == nil {
		return
	}

	pos := player.GetPosition()
	tile := pos.Tile()
	centerX, centerY := int(tile.X()), int(tile.Y())

	for y := centerY - miniMapExploreRadiusTiles; y <= centerY+miniMapExploreRadiusTiles; y++ {
		for x := centerX - miniMapExploreRadiusTiles; x <= centerX+miniMapExploreRadiusTiles; x++ {
			mapEngine.MarkExplored(x, y)
		}
	}
}

// miniMapWindow returns the [minX,maxX) x [minY,maxY) tile rectangle a
// MiniMap draws around (playerX, playerY), clamped to the map's own
// bounds (mapWidth, mapHeight) -- extracted so the clamping logic is
// testable without a live d2mapengine.MapEngine/d2interface.Surface, same
// reason as this session's other pure-logic extractions (amplifiedDamage,
// overloadBonusDamage, shouldIgniteBurn). Deliberately a flat tile
// rectangle, not an isometric viewport: d2maprenderer's own
// WorldToOrtho/OrthoToScreen (viewport.go) do isometric skew a top-down
// minimap doesn't want.
func miniMapWindow(playerX, playerY, mapWidth, mapHeight int) (minX, minY, maxX, maxY int) {
	minX = playerX - miniMapWindowRadiusTiles
	minY = playerY - miniMapWindowRadiusTiles
	maxX = playerX + miniMapWindowRadiusTiles
	maxY = playerY + miniMapWindowRadiusTiles

	if minX < 0 {
		minX = 0
	}

	if minY < 0 {
		minY = 0
	}

	if maxX > mapWidth {
		maxX = mapWidth
	}

	if maxY > mapHeight {
		maxY = mapHeight
	}

	return minX, minY, maxX, maxY
}

// Render draws a small top-left corner overlay: a translucent background,
// one filled square per explored+walkable tile within miniMapWindow, a
// distinct marker for the player's own current tile, and a larger distinct
// marker for each explored waypoint (see d2mapentity.WaypointTiles) --
// devil_game_design_reference.md §4.1's "icônes de points d'intérêt",
// waypoints only for now. A no-op if not Visible, or if mapEngine/player
// is nil.
func (m *MiniMap) Render(target d2interface.Surface, mapEngine *d2mapengine.MapEngine, player *d2mapentity.Player) {
	if !m.Visible || mapEngine == nil || player == nil {
		return
	}

	pos := player.GetPosition()
	tile := pos.Tile()
	playerX, playerY := int(tile.X()), int(tile.Y())

	size := mapEngine.Size()
	minX, minY, maxX, maxY := miniMapWindow(playerX, playerY, size.Width, size.Height)

	target.PushTranslation(miniMapMarginX, miniMapMarginY)
	defer target.Pop()

	windowWidth := (maxX - minX) * miniMapTileScale
	windowHeight := (maxY - minY) * miniMapTileScale
	target.DrawRect(windowWidth, windowHeight, d2util.Color(miniMapBackgroundColor))

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			if !mapEngine.IsExplored(x, y) {
				continue
			}

			mapTile := mapEngine.TileAt(x, y)
			if mapTile == nil {
				continue
			}

			if mapTile.GetSubTileFlags(miniMapCenterSubtile, miniMapCenterSubtile).BlockWalk {
				continue // skip walls entirely -- keep the overlay to walkable floor only, same minimal scope as the rest of this file
			}

			dotColor := d2util.Color(miniMapExploredColor)
			if x == playerX && y == playerY {
				dotColor = d2util.Color(miniMapPlayerColor)
			}

			target.PushTranslation((x-minX)*miniMapTileScale, (y-minY)*miniMapTileScale)
			target.DrawRect(miniMapTileScale, miniMapTileScale, dotColor)
			target.Pop()
		}
	}

	m.renderWaypoints(target, mapEngine, minX, minY, maxX, maxY)
}

// renderWaypoints draws a larger, distinct marker for each explored
// waypoint within [minX,maxX) x [minY,maxY) -- devil_game_design_reference.md
// §4.1's "icônes de points d'intérêt", waypoints only for now (Boss/
// rare-chest markers aren't included: no boss placement or a rare-chest
// identification mechanism exists yet, ROADMAP.md). Split out of Render
// to keep its own cyclomatic complexity down, same practice as this
// session's other extractions.
//
// Only explored waypoints are drawn ("dès leur découverte") -- one
// sitting in an unexplored part of the level shouldn't leak through the
// overlay before the player has actually found it.
func (m *MiniMap) renderWaypoints(target d2interface.Surface, mapEngine *d2mapengine.MapEngine, minX, minY, maxX, maxY int) {
	for _, waypointTile := range d2mapentity.WaypointTiles(mapEngine.Entities()) {
		x, y := waypointTile[0], waypointTile[1]
		if x < minX || x >= maxX || y < minY || y >= maxY {
			continue
		}

		if !mapEngine.IsExplored(x, y) {
			continue
		}

		target.PushTranslation((x-minX)*miniMapTileScale, (y-minY)*miniMapTileScale)
		target.DrawRect(miniMapWaypointScale, miniMapWaypointScale, d2util.Color(miniMapWaypointColor))
		target.Pop()
	}
}
