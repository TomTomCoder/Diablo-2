package d2player

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2geom"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapengine"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2maprenderer"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

const (
	logPrefix = "Player"
)

// Panel represents the panel at the bottom of the game screen
type Panel interface {
	IsOpen() bool
	Open()
	Close()
}

const mouseBtnActionsThreshold = 0.25

const (
	// Since they require special handling, not considering (1) globes, (2) content of the mini panel, (3) belt
	leftSkill actionableType = iota
	xp
	stamina
	rightSkill
	hpGlobe
	manaGlobe
)

// sorceressInventoryRecordKey is Sorceress2's InventoryRecord key --
// HeroDevil's own fallback below, no InventoryRecord of its own existing.
const sorceressInventoryRecordKey = "Sorceress2"

const (
	leftSkillX,
	leftSkillY,
	leftSkillWidth,
	leftSkillHeight = 117, 550, 50, 50

	xpX,
	xpY,
	xpWidth,
	xpHeight = 253, 560, 125, 5

	staminaX,
	staminaY,
	staminaWidth,
	staminaHeight = 273, 573, 105, 20

	rightSkillX,
	rightSkillY,
	rightSkillWidth,
	rightSkillHeight = 635, 550, 50, 50

	hpGlobeX,
	hpGlobeY,
	hpGlobeWidth,
	hpGlobeHeight = 30, 525, 80, 60

	manaGlobeX,
	manaGlobeY,
	manaGlobeWidth,
	manaGlobeHeight = 695, 525, 80, 60
)

const (
	menuBottomRectX,
	menuBottomRectY,
	menuBottomRectW,
	menuBottomRectH = 0, 550, 800, 50

	menuLeftRectX,
	menuLeftRectY,
	menuLeftRectW,
	menuLeftRectH = 0, 0, 400, 600

	menuRightRectX,
	menuRightRectY,
	menuRightRectW,
	menuRightRectH = 400, 0, 400, 600
)

// NewGameControls creates a GameControls instance and returns a pointer to it
// nolint:funlen // doesn't make sense to split this up
func NewGameControls(
	asset *d2asset.AssetManager,
	renderer d2interface.Renderer,
	hero *d2mapentity.Player,
	mapEngine *d2mapengine.MapEngine,
	escapeMenu *EscapeMenu,
	mapRenderer *d2maprenderer.MapRenderer,
	inputListener inputCallbackListener,
	term d2interface.Terminal,
	ui *d2ui.UIManager,
	keyMap *KeyMap,
	audioProvider d2interface.AudioProvider,
	l d2util.LogLevel,
	isSinglePlayer bool,
	players map[string]*d2mapentity.Player,
) (*GameControls, error) {
	var inventoryRecordKey string

	switch hero.Class {
	case d2enum.HeroAssassin:
		inventoryRecordKey = "Assassin2"
	case d2enum.HeroAmazon:
		inventoryRecordKey = "Amazon2"
	case d2enum.HeroBarbarian:
		inventoryRecordKey = "Barbarian2"
	case d2enum.HeroDruid:
		inventoryRecordKey = "Druid2"
	case d2enum.HeroNecromancer:
		inventoryRecordKey = "Necromancer2"
	case d2enum.HeroPaladin:
		inventoryRecordKey = "Paladin2"
	case d2enum.HeroSorceress:
		inventoryRecordKey = sorceressInventoryRecordKey
	case d2enum.HeroDevil:
		// ponytail: no HeroDevil-specific InventoryRecord exists (this key
		// selects a stock inventory.txt panel layout, not something this
		// codebase generates) -- reusing Sorceress2's layout as the
		// closest thematic fit (a caster class) rather than inventing a
		// new layout, same practice as Devil's own skill icons falling
		// back to the generic sheet (NewDevilHeroSkill's Charclass "").
		inventoryRecordKey = sorceressInventoryRecordKey
	default:
		return nil, fmt.Errorf("unknown hero class: %d", hero.Class)
	}

	actionableRegions := []actionableRegion{
		{leftSkill, d2geom.Rectangle{
			Left:   leftSkillX,
			Top:    leftSkillY,
			Width:  leftSkillWidth,
			Height: leftSkillHeight,
		}},
		{xp, d2geom.Rectangle{
			Left:   xpX,
			Top:    xpY,
			Width:  xpWidth,
			Height: xpHeight,
		}},
		{stamina, d2geom.Rectangle{
			Left:   staminaX,
			Top:    staminaY,
			Width:  staminaWidth,
			Height: staminaHeight,
		}},
		{rightSkill, d2geom.Rectangle{
			Left:   rightSkillX,
			Top:    rightSkillY,
			Width:  rightSkillWidth,
			Height: rightSkillHeight,
		}},
		{hpGlobe, d2geom.Rectangle{
			Left:   hpGlobeX,
			Top:    hpGlobeY,
			Width:  hpGlobeWidth,
			Height: hpGlobeHeight,
		}},
		{manaGlobe, d2geom.Rectangle{
			Left:   manaGlobeX,
			Top:    manaGlobeY,
			Width:  manaGlobeWidth,
			Height: manaGlobeHeight,
		}},
	}
	inventoryRecord := asset.Records.Layout.Inventory[inventoryRecordKey]

	heroStatsPanel := NewHeroStatsPanel(asset, ui, hero.Name(), hero.Class, l, hero.Stats)

	questLog := NewQuestLog(asset, ui, l, audioProvider, hero.Act)

	inventory, err := NewInventory(asset, ui, l, hero.Gold, inventoryRecord)
	if err != nil {
		return nil, err
	}

	skilltree := newSkillTree(hero.Skills, hero.Class, hero.Stats, asset, l, ui)

	devilInventoryPanel := newDevilInventoryPanel(ui, l)

	miniPanel := newMiniPanel(asset, ui, l, isSinglePlayer)

	heroState, err := d2hero.NewHeroStateFactory(asset)
	if err != nil {
		return nil, err
	}

	helpOverlay := NewHelpOverlay(asset, ui, l, keyMap)

	const blackAlpha50percent = 0x0000007f

	gc := &GameControls{
		asset:               asset,
		ui:                  ui,
		renderer:            renderer,
		hero:                hero,
		heroState:           heroState,
		escapeMenu:          escapeMenu,
		inputListener:       inputListener,
		mapRenderer:         mapRenderer,
		inventory:           inventory,
		skilltree:           skilltree,
		devilInventoryPanel: devilInventoryPanel,
		heroStatsPanel:      heroStatsPanel,
		questLog:            questLog,
		HelpOverlay:         helpOverlay,
		keyMap:              keyMap,
		bottomMenuRect: &d2geom.Rectangle{
			Left:   menuBottomRectX,
			Top:    menuBottomRectY,
			Width:  menuBottomRectW,
			Height: menuBottomRectH,
		},
		leftMenuRect: &d2geom.Rectangle{
			Left:   menuLeftRectX,
			Top:    menuLeftRectY,
			Width:  menuLeftRectW,
			Height: menuLeftRectH,
		},
		rightMenuRect: &d2geom.Rectangle{
			Left:   menuRightRectX,
			Top:    menuRightRectY,
			Width:  menuRightRectW,
			Height: menuRightRectH,
		},
		actionableRegions:      actionableRegions,
		lastLeftBtnActionTime:  0,
		lastRightBtnActionTime: 0,
		isSinglePlayer:         isSinglePlayer,
	}

	if !isSinglePlayer {
		PartyPanel := NewPartyPanel(asset, ui, hero.Name(), l, hero, hero.Stats, players)
		gc.PartyPanel = PartyPanel
	}

	hud := NewHUD(asset, ui, hero, miniPanel, actionableRegions, mapEngine, l, gc, mapRenderer)
	gc.hud = hud

	hoverLabel := hud.nameLabel
	hoverLabel.SetBackgroundColor(d2util.Color(blackAlpha50percent))

	gc.heroStatsPanel.SetOnCloseCb(gc.onCloseHeroStatsPanel)
	gc.heroStatsPanel.SetOnSpendPointCb(gc.inputListener.OnSpendAttributePoint)
	gc.heroStatsPanel.SetOnForgetAttributeCb(gc.inputListener.OnRespecSingleAttributePoint)
	gc.questLog.SetOnCloseCb(gc.onCloseQuestLog)
	gc.inventory.SetOnCloseCb(gc.onCloseInventory)
	gc.skilltree.SetOnCloseCb(gc.onCloseSkilltree)
	gc.skilltree.SetOnInvestSkillPointCb(gc.inputListener.OnInvestSkillPoint)
	gc.skilltree.SetOnLearnSkillCb(gc.inputListener.OnLearnSkill)
	gc.skilltree.SetOnForgetSkillCb(gc.inputListener.OnRespecSingleSkill)
	gc.devilInventoryPanel.SetOnCloseCb(gc.onCloseDevilInventoryPanel)
	gc.devilInventoryPanel.SetOnMoveToStashCb(gc.inputListener.OnMoveToStash)
	gc.devilInventoryPanel.SetOnMoveToInventoryCb(gc.inputListener.OnMoveToInventory)
	gc.devilInventoryPanel.SetOnMoveFromBeltCb(gc.inputListener.OnMoveFromBelt)
	gc.hud.skillSelectMenu.SetOnEquipCb(gc.inputListener.OnEquipSkill)

	gc.escapeMenu.SetOnCloseCb(gc.hud.miniPanel.restoreDisabled)
	gc.HelpOverlay.SetOnCloseCb(gc.hud.miniPanel.restoreDisabled)

	err = gc.bindTerminalCommands(term)
	if err != nil {
		return nil, err
	}

	gc.Logger = d2util.NewLogger()
	gc.Logger.SetLevel(l)
	gc.Logger.SetPrefix(logPrefix)

	return gc, nil
}

// GameControls represents the game's controls on the screen
type GameControls struct {
	keyMap                 *KeyMap
	actionableRegions      []actionableRegion
	asset                  *d2asset.AssetManager
	renderer               d2interface.Renderer // https://github.com/OpenDiablo2/OpenDiablo2/issues/798
	inputListener          inputCallbackListener
	hero                   *d2mapentity.Player
	heroState              *d2hero.HeroStateFactory
	mapRenderer            *d2maprenderer.MapRenderer
	escapeMenu             *EscapeMenu
	ui                     *d2ui.UIManager
	inventory              *Inventory
	hud                    *HUD
	skilltree              *skillTree
	devilInventoryPanel    *devilInventoryPanel
	heroStatsPanel         *HeroStatsPanel
	PartyPanel             *PartyPanel
	questLog               *QuestLog
	HelpOverlay            *HelpOverlay
	bottomMenuRect         *d2geom.Rectangle
	leftMenuRect           *d2geom.Rectangle
	rightMenuRect          *d2geom.Rectangle
	lastMouseX             int
	lastMouseY             int
	lastLeftBtnActionTime  float64
	lastRightBtnActionTime float64
	FreeCam                bool
	isSinglePlayer         bool

	*d2util.Logger
}

type actionableType int

type actionableRegion struct {
	actionableTypeID actionableType
	rect             d2geom.Rectangle
}

// SkillResource represents a Skill with its corresponding icon sprite, path to DC6 file and icon number.
// SkillResourcePath points to a DC6 resource which contains the icons of multiple skills as frames.
// The IconNumber is the frame at which we can find our skill sprite in the DC6 file.
type SkillResource struct {
	SkillResourcePath string // path to a skills DC6 file(see getSkillResourceByClass)
	IconNumber        int    // the index of the frame in the DC6 file
	SkillIcon         *d2ui.Sprite
}

// OnKeyRepeat is called to handle repeated key presses
func (g *GameControls) OnKeyRepeat(event d2interface.KeyEvent) bool {
	if g.FreeCam {
		var moveSpeed float64 = 8
		if event.KeyMod() == d2enum.KeyModShift {
			moveSpeed *= 2
		}

		if event.Key() == d2enum.KeyDown {
			v := d2vector.NewVector(0, moveSpeed)
			g.mapRenderer.MoveCameraTargetBy(v)

			return true
		}

		if event.Key() == d2enum.KeyUp {
			v := d2vector.NewVector(0, -moveSpeed)
			g.mapRenderer.MoveCameraTargetBy(v)

			return true
		}

		if event.Key() == d2enum.KeyRight {
			v := d2vector.NewVector(moveSpeed, 0)
			g.mapRenderer.MoveCameraTargetBy(v)

			return true
		}

		if event.Key() == d2enum.KeyLeft {
			v := d2vector.NewVector(-moveSpeed, 0)
			g.mapRenderer.MoveCameraTargetBy(v)

			return true
		}
	}

	return false
}

// OnKeyDown handles key presses -- a flat one-case-per-keybinding dispatch
// switch, not real branching complexity, same practice as
// NewGameControls' own nolint:funlen just above.
// nolint:gocyclo // flat dispatch switch, not real complexity
func (g *GameControls) OnKeyDown(event d2interface.KeyEvent) bool {
	if event.Key() == d2enum.KeyEscape {
		g.onEscKey()
		return true
	}

	gameEvent := g.keyMap.getGameEvent(event.Key())

	switch gameEvent {
	case d2enum.ClearScreen:
		g.clearScreen()
		g.updateLayout()
	case d2enum.ToggleInventoryPanel:
		g.toggleInventoryPanel()
	case d2enum.TogglePartyPanel:
		if !g.isSinglePlayer {
			g.togglePartyPanel()
		}
	case d2enum.ToggleSkillTreePanel:
		g.toggleSkilltreePanel()
	case d2enum.ToggleCharacterPanel:
		g.toggleHeroStatsPanel()
	case d2enum.ToggleQuestLog:
		g.toggleQuestLog()
	case d2enum.ToggleRunWalk:
		g.hud.onToggleRunButton(false)
	case d2enum.HoldRun:
		g.hud.onToggleRunButton(true)
	case d2enum.ToggleHelpScreen:
		g.toggleHelpOverlay()
	case d2enum.CraftItem:
		g.inputListener.OnCraft(d2hero.RecipeUpgradeBatonApprenti)
	case d2enum.ToggleOverload:
		g.inputListener.OnToggleOverload()
	case d2enum.ToggleDevilInventory:
		g.toggleDevilInventoryPanel()
	case d2enum.UseRespecEssence:
		g.inputListener.OnUseRespecEssence()
	case d2enum.ArmGlypheDOubli:
		// Arms both panels: the player doesn't pre-commit to
		// forgetting a skill vs. an attribute point when pressing this
		// key, only when they click a specific icon/button afterwards.
		// Clicking one disarms only that panel's own flag -- the other
		// stays armed harmlessly until its own next click or a repeat
		// of this keybinding, since neither panel shares state with
		// the other.
		g.skilltree.ArmGlypheDOubli()
		g.heroStatsPanel.ArmGlypheDOubli()
	default:
		return false
	}

	return false
}

// OnKeyUp handles key release
func (g *GameControls) OnKeyUp(event d2interface.KeyEvent) bool {
	gameEvent := g.keyMap.getGameEvent(event.Key())

	if gameEvent == d2enum.HoldRun {
		g.hud.onToggleRunButton(true)
	}

	return false
}

// When escape is pressed:
// 1. If there was some overlay or panel open, close it
// 2. Otherwise, if the Escape Menu was open, let the Escape Menu handle it
// 3. If nothing was open, open the Escape Menu
func (g *GameControls) onEscKey() {
	escHandled := false

	escHandled = g.hasOpenPanels() || g.HelpOverlay.IsOpen() || g.hud.skillSelectMenu.IsOpen()
	g.clearScreen()

	if escHandled {
		g.updateLayout()
		return
	}

	if g.escapeMenu.IsOpen() {
		g.escapeMenu.OnEscKey()
	} else {
		g.openEscMenu()
	}
}

func truncateFloat64(n float64) float64 {
	const ten = 10.0
	return float64(int(n*ten)) / ten
}

// OnMouseButtonRepeat handles repeated mouse clicks
func (g *GameControls) OnMouseButtonRepeat(event d2interface.MouseEvent) bool {
	const (
		screenWidth, screenHeight         = 800, 600
		halfScreenWidth, halfScreenHeight = screenWidth / 2, screenHeight / 2
		subtilesPerTile                   = 5
	)

	px, py := g.mapRenderer.ScreenToWorld(event.X(), event.Y())
	px = truncateFloat64(px)
	py = truncateFloat64(py)

	now := d2util.Now()
	button := event.Button()
	isLeft := button == d2enum.MouseButtonLeft
	isRight := button == d2enum.MouseButtonRight
	lastLeft := now - g.lastLeftBtnActionTime
	lastRight := now - g.lastRightBtnActionTime
	inRect := !g.isInActiveMenusRect(event.X(), event.Y())
	shouldDoLeft := lastLeft >= mouseBtnActionsThreshold
	shouldDoRight := lastRight >= mouseBtnActionsThreshold

	if isLeft && shouldDoLeft && inRect && !g.hero.IsCasting() {
		g.lastLeftBtnActionTime = now

		if event.KeyMod() == d2enum.KeyModShift {
			if id, ok := castSkillID(g.hero.LeftSkill); ok {
				g.inputListener.OnPlayerCast(id, px, py)
			}
		} else {
			g.inputListener.OnPlayerMove(px, py)
		}

		if g.FreeCam {
			if event.Button() == d2enum.MouseButtonLeft {
				camVect := g.mapRenderer.Camera.GetPosition().Vector

				x := float64(halfScreenWidth) / subtilesPerTile
				y := float64(halfScreenHeight) / subtilesPerTile

				targetPosition := d2vector.NewPositionTile(x, y)
				targetPosition.Add(&camVect)

				g.mapRenderer.SetCameraTarget(&targetPosition)

				return true
			}
		}

		return true
	}

	if isRight && shouldDoRight && inRect && !g.hero.IsCasting() {
		g.lastRightBtnActionTime = now

		if id, ok := castSkillID(g.hero.RightSkill); ok {
			g.inputListener.OnPlayerCast(id, px, py)
		}

		return true
	}

	return true
}

// OnMouseMove handles mouse movement events
func (g *GameControls) OnMouseMove(event d2interface.MouseMoveEvent) bool {
	mx, my := event.X(), event.Y()
	g.lastMouseX = mx
	g.lastMouseY = my
	g.inventory.lastMouseX = mx
	g.inventory.lastMouseY = my

	for i := range g.actionableRegions {
		// Mouse over a game control element
		if g.actionableRegions[i].rect.IsInRect(mx, my) {
			g.onHoverActionable(g.actionableRegions[i].actionableTypeID)
		}
	}

	g.hud.OnMouseMove(event)

	if g.PartyPanel != nil {
		g.PartyPanel.OnMouseMove(event)
	}

	return false
}

// OnMouseButtonUp handles mouse button presses
func (g *GameControls) OnMouseButtonUp(event d2interface.MouseEvent) bool {
	return false
}

// OnMouseButtonDown handles mouse button presses
func (g *GameControls) OnMouseButtonDown(event d2interface.MouseEvent) bool {
	mx, my := event.X(), event.Y()

	for i := range g.actionableRegions {
		// If click is on a game control element
		if g.actionableRegions[i].rect.IsInRect(mx, my) {
			g.onClickActionable(g.actionableRegions[i].actionableTypeID)
			return false
		}
	}

	if g.hud.skillSelectMenu.IsOpen() && event.Button() == d2enum.MouseButtonLeft {
		g.lastLeftBtnActionTime = d2util.Now()
		g.hud.skillSelectMenu.HandleClick(mx, my)
		g.hud.skillSelectMenu.ClosePanels()

		return false
	}

	px, py := g.mapRenderer.ScreenToWorld(mx, my)
	px = truncateFloat64(px)
	py = truncateFloat64(py)

	if event.Button() == d2enum.MouseButtonLeft && !g.isInActiveMenusRect(mx, my) && !g.hero.IsCasting() {
		g.lastLeftBtnActionTime = d2util.Now()

		if event.KeyMod() == d2enum.KeyModShift {
			if id, ok := castSkillID(g.hero.LeftSkill); ok {
				g.inputListener.OnPlayerCast(id, px, py)
			}
		} else {
			g.inputListener.OnPlayerMove(px, py)
		}

		return true
	}

	if event.Button() == d2enum.MouseButtonRight && !g.isInActiveMenusRect(mx, my) && !g.hero.IsCasting() {
		g.lastRightBtnActionTime = d2util.Now()

		if id, ok := castSkillID(g.hero.RightSkill); ok {
			g.inputListener.OnPlayerCast(id, px, py)
		}

		return true
	}

	return false
}

func (g *GameControls) clearLeftScreenSide() {
	g.heroStatsPanel.Close()

	if g.PartyPanel != nil {
		g.PartyPanel.Close()
	}

	g.questLog.Close()
	g.hud.skillSelectMenu.ClosePanels()
	g.updateLayout()
}

func (g *GameControls) clearRightScreenSide() {
	g.inventory.Close()
	g.skilltree.Close()
	g.devilInventoryPanel.Close()
	g.hud.skillSelectMenu.ClosePanels()
	g.updateLayout()
}

func (g *GameControls) clearScreen() {
	g.clearRightScreenSide()
	g.clearLeftScreenSide()
	g.hud.skillSelectMenu.ClosePanels()
	g.HelpOverlay.Close()
}

func (g *GameControls) openLeftPanel(panel Panel) {
	if !g.HelpOverlay.IsOpen() && !g.escapeMenu.IsOpen() {
		isOpen := panel.IsOpen()

		g.clearLeftScreenSide()

		if !isOpen {
			panel.Open()
			g.updateLayout()
		}
	}
}

func (g *GameControls) openRightPanel(panel Panel) {
	if !g.HelpOverlay.IsOpen() && !g.escapeMenu.IsOpen() {
		isOpen := panel.IsOpen()

		g.clearRightScreenSide()

		if !isOpen {
			panel.Open()
			g.updateLayout()
		}
	}
}

func (g *GameControls) toggleHeroStatsPanel() {
	g.openLeftPanel(g.heroStatsPanel)
}

func (g *GameControls) togglePartyPanel() {
	g.openLeftPanel(g.PartyPanel)
}

func (g *GameControls) onCloseHeroStatsPanel() {
	g.updateLayout()
}

func (g *GameControls) toggleLeftSkillPanel() {
	if !g.HelpOverlay.IsOpen() {
		g.clearScreen()
		g.hud.skillSelectMenu.ToggleLeftPanel()
	}
}

func (g *GameControls) toggleRightSkillPanel() {
	if !g.HelpOverlay.IsOpen() {
		g.clearScreen()
		g.hud.skillSelectMenu.ToggleRightPanel()
	}
}

func (g *GameControls) toggleQuestLog() {
	g.openLeftPanel(g.questLog)
}

func (g *GameControls) onCloseQuestLog() {
	g.updateLayout()
}

func (g *GameControls) toggleHelpOverlay() {
	if !g.isRightPanelOpen() || g.isLeftPanelOpen() {
		g.HelpOverlay.updateKeyMap(g.keyMap)
		g.hud.skillSelectMenu.ClosePanels()
		g.hud.miniPanel.openDisabled()
		g.HelpOverlay.Toggle()
		g.updateLayout()
	}
}

func (g *GameControls) toggleInventoryPanel() {
	g.openRightPanel(g.inventory)
}

func (g *GameControls) onCloseInventory() {
	g.updateLayout()
}

func (g *GameControls) toggleSkilltreePanel() {
	g.openRightPanel(g.skilltree)
}

func (g *GameControls) onCloseSkilltree() {
	g.updateLayout()
}

func (g *GameControls) toggleDevilInventoryPanel() {
	g.openRightPanel(g.devilInventoryPanel)
}

func (g *GameControls) onCloseDevilInventoryPanel() {
	g.updateLayout()
}

func (g *GameControls) openEscMenu() {
	g.clearScreen()
	g.hud.miniPanel.closeDisabled()
	g.escapeMenu.open()
	g.updateLayout()
}

// Load the resources required for the GameControls
func (g *GameControls) Load() {
	g.hud.Load()
	g.inventory.Load()
	g.skilltree.load()
	g.devilInventoryPanel.load(g.hero)
	g.heroStatsPanel.Load()

	if g.PartyPanel != nil {
		g.PartyPanel.Load()
	}

	g.questLog.Load()
	g.HelpOverlay.Load()

	g.loadAddButtons()
	g.setAddButtons()

	miniPanelActions := &miniPanelActions{
		characterToggle: g.toggleHeroStatsPanel,
		partyToggle:     g.togglePartyPanel,
		inventoryToggle: g.toggleInventoryPanel,
		skilltreeToggle: g.toggleSkilltreePanel,
		menuToggle:      g.openEscMenu,
		questToggle:     g.toggleQuestLog,
	}
	g.hud.miniPanel.load(miniPanelActions)
}

// Advance advances the state of the GameControls
func (g *GameControls) Advance(elapsed float64) error {
	g.mapRenderer.Advance(elapsed)
	g.hud.Advance(elapsed)
	g.inventory.Advance(elapsed)
	g.questLog.Advance(elapsed)

	if g.PartyPanel != nil {
		g.PartyPanel.Advance(elapsed)
	}

	if err := g.escapeMenu.Advance(elapsed); err != nil {
		return err
	}

	if g.heroStatsPanel.IsOpen() || g.skilltree.IsOpen() {
		g.setAddButtons()
	}

	return nil
}

func (g *GameControls) updateLayout() {
	isRightPanelOpen := g.isLeftPanelOpen()
	isLeftPanelOpen := g.isRightPanelOpen()

	switch {
	case isRightPanelOpen == isLeftPanelOpen:
		g.hud.miniPanel.ResetPosition()
		g.mapRenderer.ViewportDefault()
	case isRightPanelOpen:
		g.hud.miniPanel.SetMovedRight(true)
		g.mapRenderer.ViewportToLeft()
	case isLeftPanelOpen:
		g.hud.miniPanel.SetMovedLeft(true)
		g.mapRenderer.ViewportToRight()
	}
}

func (g *GameControls) isLeftPanelOpen() bool {
	var partyPanel bool

	if g.PartyPanel != nil {
		partyPanel = g.PartyPanel.IsOpen()
	} else {
		partyPanel = false
	}

	return g.heroStatsPanel.IsOpen() || partyPanel || g.questLog.IsOpen() || g.inventory.moveGoldPanel.IsOpen()
}

func (g *GameControls) isRightPanelOpen() bool {
	return g.inventory.IsOpen() || g.skilltree.IsOpen()
}

func (g *GameControls) hasOpenPanels() bool {
	return g.isRightPanelOpen() || g.isLeftPanelOpen() || g.hud.skillSelectMenu.IsOpen()
}

func (g *GameControls) isInActiveMenusRect(px, py int) bool {
	if g.bottomMenuRect.IsInRect(px, py) {
		return true
	}

	if g.isLeftPanelOpen() && g.leftMenuRect.IsInRect(px, py) {
		return true
	}

	if g.isRightPanelOpen() && g.rightMenuRect.IsInRect(px, py) {
		return true
	}

	if g.hud.miniPanel.IsOpen() && g.hud.miniPanel.IsInRect(px, py) {
		return true
	}

	if g.escapeMenu.IsOpen() {
		return true
	}

	if g.HelpOverlay.IsOpen() && g.HelpOverlay.IsInRect(px, py) {
		return true
	}

	if g.hud.skillSelectMenu.IsOpen() {
		return true
	}

	return false
}

// Render draws the GameControls onto the target
func (g *GameControls) Render(target d2interface.Surface) error {
	if err := g.hud.Render(target); err != nil {
		return err
	}

	if err := g.renderPanels(target); err != nil {
		return err
	}

	if err := g.escapeMenu.Render(target); err != nil {
		return err
	}

	return nil
}

func (g *GameControls) renderPanels(target d2interface.Surface) error {
	g.inventory.Render(target)

	return nil
}

// SetZoneChangeText sets the zoneChangeText
func (g *GameControls) SetZoneChangeText(text string) {
	g.hud.zoneChangeText.SetText(text)
}

// ShowZoneChangeText shows the zoneChangeText
func (g *GameControls) ShowZoneChangeText() {
	g.hud.isZoneTextShown = true
}

// HideZoneChangeTextAfter hides the zoneChangeText after the given amount of seconds
func (g *GameControls) HideZoneChangeTextAfter(delay float64) {
	time.AfterFunc(time.Duration(delay)*time.Second, func() {
		g.hud.isZoneTextShown = false
	})
}

// HpStatsIsVisible returns true if the hp and mana stats are visible to the player
func (g *GameControls) HpStatsIsVisible() bool {
	return g.hud.hpStatsIsVisible
}

// ManaStatsIsVisible returns true if the hp and mana stats are visible to the player
func (g *GameControls) ManaStatsIsVisible() bool {
	return g.hud.manaStatsIsVisible
}

// ToggleHpStats toggles the visibility of the hp and mana stats placed above their respective globe and load only if they do not match
func (g *GameControls) ToggleHpStats() {
	g.hud.hpStatsIsVisible = !g.hud.hpStatsIsVisible
}

// ToggleManaStats toggles the visibility of the hp and mana stats placed above their respective globe
func (g *GameControls) ToggleManaStats() {
	g.hud.manaStatsIsVisible = !g.hud.manaStatsIsVisible
}

// Handles what to do when an actionable is hovered
func (g *GameControls) onHoverActionable(item actionableType) {
	hoverMap := map[actionableType]func(){
		leftSkill:  func() {},
		xp:         func() {},
		stamina:    func() {},
		rightSkill: func() {},
		hpGlobe:    func() {},
		manaGlobe:  func() {},
	}

	onHover, found := hoverMap[item]
	if !found {
		g.Errorf("Unrecognized actionableType(%d) being hovered", item)
		return
	}

	onHover()
}

// Handles what to do when an actionable is clicked
func (g *GameControls) onClickActionable(item actionableType) {
	actionMap := map[actionableType]func(){
		leftSkill: func() {
			g.toggleLeftSkillPanel()
		},

		xp: func() {
			g.Info("XP Action Pressed")
		},

		stamina: func() {
			g.Info("Stamina Action Pressed")
		},

		rightSkill: func() {
			g.toggleRightSkillPanel()
		},

		hpGlobe: func() {
			g.ToggleHpStats()
			g.Info("HP Globe Pressed")
		},

		manaGlobe: func() {
			g.ToggleManaStats()
			g.Info("Mana Globe Pressed")
		},
	}

	action, found := actionMap[item]
	if !found {
		// Warning, because some action types are still todo, and could return this error
		g.Warningf("Unrecognized actionableType(%d) being clicked", item)
		return
	}

	action()
}

func (g *GameControls) bindTerminalCommands(term d2interface.Terminal) error {
	if err := term.Bind("freecam", "toggle free camera movement", nil, g.commandFreeCam); err != nil {
		return err
	}

	if err := term.Bind("setleftskill", "set skill to fire on left click", []string{"id"}, g.commandSetLeftSkill(term)); err != nil {
		return err
	}

	if err := term.Bind("setrightskill", "set skill to fire on right click", []string{"id"}, g.commandSetRightSkill(term)); err != nil {
		return err
	}

	if err := term.Bind("learnskills", "learn all skills for the a given class", []string{"token"}, g.commandLearnSkills(term)); err != nil {
		return err
	}

	if err := term.Bind("learnskillid", "learn a skill by a given ID", []string{"id"}, g.commandLearnSkillID(term)); err != nil {
		return err
	}

	return nil
}

// UnbindTerminalCommands unbinds commands from the terminal
func (g *GameControls) UnbindTerminalCommands(term d2interface.Terminal) error {
	return term.Unbind("freecam", "setleftskill", "setrightskill", "learnskills", "learnskillid")
}

func (g *GameControls) setAddButtons() {
	g.hud.addStatsButton.SetEnabled(g.hero.Stats.StatsPoints > 0)
	g.hud.addSkillButton.SetEnabled(g.hero.Stats.SkillPoints > 0)
}

func (g *GameControls) loadAddButtons() {
	g.hud.addStatsButton.OnActivated(func() { g.toggleHeroStatsPanel() })
	g.hud.addSkillButton.OnActivated(func() { g.toggleSkilltreePanel() })
}

func (g *GameControls) commandFreeCam([]string) error {
	g.FreeCam = !g.FreeCam

	return nil
}

func (g *GameControls) commandSetLeftSkill(term d2interface.Terminal) func(args []string) error {
	return func(args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			term.Errorf("invalid argument")
			return nil
		}

		skill, err := g.heroSkillByID(id)
		if err != nil {
			term.Errorf(err.Error())
			return nil
		}

		g.hero.LeftSkill = skill

		return nil
	}
}

func (g *GameControls) commandSetRightSkill(term d2interface.Terminal) func(args []string) error {
	return func(args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			term.Errorf("invalid argument")
			return nil
		}

		skill, err := g.heroSkillByID(id)
		if err != nil {
			term.Errorf(err.Error())
			return nil
		}

		g.hero.RightSkill = skill

		return nil
	}
}

func (g *GameControls) commandLearnSkillID(term d2interface.Terminal) func(args []string) error {
	return func(args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			term.Errorf("invalid argument")
			return nil
		}

		skill, err := g.heroSkillByID(id)
		if err != nil {
			term.Errorf("%s", err.Error())
			return nil
		}

		g.hero.Skills[skill.ID] = skill
		g.hud.skillSelectMenu.RegenerateImageCache()
		g.Infof("Learned skill: %s", skill.Skill)

		return nil
	}
}

func (g *GameControls) heroSkillByID(id int) (*d2hero.HeroSkill, error) {
	skillRecord := g.asset.Records.Skill.Details[id]
	if skillRecord == nil {
		return nil, fmt.Errorf("cannot find a skill record for ID: %d", id)
	}

	skill, err := g.heroState.CreateHeroSkill(1, skillRecord.Skill)
	if err != nil {
		return nil, fmt.Errorf("cannot create skill with ID of %d", id)
	}

	return skill, nil
}

// castSkillID returns skill.ID and true, or (0, false) if no skill is
// assigned. Guards OnPlayerCast callers against a nil LeftSkill/RightSkill
// -- e.g. before Devil's own skill selection assigns one (ROADMAP.md
// Phase 2), a fresh RightSkill is nil by default.
func castSkillID(skill *d2hero.HeroSkill) (id int, ok bool) {
	if skill == nil {
		return 0, false
	}

	return skill.ID, true
}

func (g *GameControls) commandLearnSkills(term d2interface.Terminal) func(args []string) error {
	const classTokenLength = 3

	return func(args []string) error {
		token := args[0]
		if len(token) < classTokenLength {
			term.Errorf("The given class token should be at least 3 characters")
			return nil
		}

		validPrefixes := []string{"ama", "ass", "nec", "bar", "sor", "dru", "pal"}
		classToken := strings.ToLower(token)
		tokenPrefix := classToken[0:3]
		isValidToken := false

		for idx := range validPrefixes {
			if strings.Compare(tokenPrefix, validPrefixes[idx]) == 0 {
				isValidToken = true
			}
		}

		if !isValidToken {
			fmtInvalid := "Invalid class, must be a value starting with(case insensitive): %s"
			term.Errorf(fmtInvalid, strings.Join(validPrefixes, ", "))

			return nil
		}

		var err error

		learnedSkillsCount := 0

		for _, skillDetailRecord := range g.asset.Records.Skill.Details {
			if skillDetailRecord.Charclass != classToken {
				continue
			}

			if skill, ok := g.hero.Skills[skillDetailRecord.ID]; ok {
				skill.SkillPoints++
				learnedSkillsCount++
			} else {
				skill, skillErr := g.heroState.CreateHeroSkill(1, skillDetailRecord.Skill)
				if skill == nil {
					continue
				}

				learnedSkillsCount++

				g.hero.Skills[skill.ID] = skill

				if skillErr != nil {
					err = skillErr
					break
				}
			}
		}

		g.hud.skillSelectMenu.RegenerateImageCache()
		g.Infof("Learned %d skills", learnedSkillsCount)

		if err != nil {
			term.Errorf("cannot learn skill for class, error: %s", err)
			return nil
		}

		return nil
	}
}
