package d2enum

// GameEvent represents an envent in the game engine
type GameEvent int

// Game events
const (
	// ToggleGameMenu will display the game menu
	ToggleGameMenu GameEvent = iota + 1

	// panel toggles
	ToggleCharacterPanel
	ToggleInventoryPanel
	TogglePartyPanel
	ToggleSkillTreePanel
	ToggleHirelingPanel
	ToggleQuestLog
	ToggleHelpScreen
	ToggleChatOverlay
	ToggleMessageLog
	ToggleRightSkillSelector // these two are for left/right speed-skill panel toggles
	ToggleLeftSkillSelector

	ToggleAutomap
	CenterAutomap        // recenters the automap when opened
	FadeAutomap          // reduces the brightness of the map (not the players/npcs)
	TogglePartyOnAutomap // toggles the display of the party members on the automap
	ToggleNamesOnAutomap // toggles the display of party members names and npcs on the automap
	ToggleMiniMap

	// there can be 16 hotkeys, each hotkey can have a skill assigned
	UseSkill1
	UseSkill2
	UseSkill3
	UseSkill4
	UseSkill5
	UseSkill6
	UseSkill7
	UseSkill8
	UseSkill9
	UseSkill10
	UseSkill11
	UseSkill12
	UseSkill13
	UseSkill14
	UseSkill15
	UseSkill16

	// switching between prev/next skill
	SelectPreviousSkill
	SelectNextSkill

	// ToggleBelts toggles the display of the different level for
	// the currently equipped belt
	ToggleBelts
	UseBeltSlot1
	UseBeltSlot2
	UseBeltSlot3
	UseBeltSlot4

	SwapWeapons
	ToggleChatBox
	ToggleRunWalk

	SayHelp
	SayFollowMe
	SayThisIsForYou
	SayThanks
	SaySorry
	SayBye
	SayNowYouDie
	SayRetreat

	// these events are fired while a player holds the corresponding key
	HoldRun
	HoldStandStill
	HoldShowGroundItems
	HoldShowPortraits

	TakeScreenShot
	ClearScreen // closes all active menus/panels
	ClearMessages

	// CraftItem attempts Devil's own single Cube de Nexus recipe
	// (d2hero.RecipeUpgradeBatonApprenti) -- ROADMAP.md Phase 5's "Craft"
	// UI item. A keybinding rather than a clickable button: with exactly
	// one recipe and no Cube de Nexus panel/art of any kind, a proper
	// recipe-picker screen would need pixel positions with no existing
	// element to safely anchor against (unlike SpendAttributePoint's
	// already-placed buttons or LearnSkill's already-gridded skill
	// icons) -- a keybinding needs none, so it's the safe way to make
	// crafting a real, triggerable player action today. Revisit once a
	// second recipe exists and picking one actually matters.
	CraftItem

	// ToggleOverload triggers Surcharge (Overload) on/off
	// (d2hero.HeroStatsState.OverloadActive). A keybinding for the same
	// reason as CraftItem -- Overload isn't one of the 30 tree skills
	// with an icon already on the skill tree grid (§6 lists it as a
	// separate generic mechanic), so there's no existing clickable
	// element to redirect the way EquipSkill/InvestSkillPoint/
	// SpendAttributePoint were.
	ToggleOverload

	// ToggleDevilInventory opens Devil's own real Inventory/Stash screen
	// (d2game/d2player's devilInventoryPanel) -- not the pre-existing
	// ToggleInventoryPanel above, which opens a panel built for stock
	// Diablo 2's spatial item-grid data model, incompatible with Devil's
	// own flat HeroState.Inventory/Stash (see devilInventoryPanel's own
	// doc comment). A new, separate keybinding rather than replacing that
	// one's target, so nothing about its existing (if data-mismatched)
	// behavior changes.
	ToggleDevilInventory

	// UseRespecEssence consumes an Essence de Boss (Ordinaire) from
	// inventory for a full respec (d2hero.HeroState.UseRespecEssence) --
	// a keybinding for the same reason as CraftItem/ToggleOverload: no
	// existing clickable element to redirect, and picking which item to
	// "use" isn't a position-based choice a screen would need anyway
	// (the server itself finds the item in inventory).
	UseRespecEssence

	// ArmGlypheDOubli arms "Glyphe d'oubli" (d2game/d2player's
	// skillTree.ArmGlypheDOubli): the next click on a known skill icon in
	// the skill tree requests forgetting it instead of investing a point.
	// A keybinding rather than a new click gesture, since the skill
	// icons' click is already claimed by invest/learn -- see
	// skillTree.ArmGlypheDOubli's own doc comment for why arming is a
	// mode rather than extending d2ui.ClickableWidget with a second click
	// type.
	ArmGlypheDOubli

	// ArmMoveToBelt arms the Inventory -> Belt direction
	// (d2game/d2player's devilInventoryPanel.ArmMoveToBelt): the next
	// click on an Inventory slot requests moving its item to the belt
	// instead of the stash. Same reasoning as ArmGlypheDOubli -- an
	// Inventory slot's click is already claimed by MoveToStash, and the
	// underlying MoveToBelt mechanism/packet/resolver already existed
	// (built alongside MoveFromBelt) but had no UI trigger until now.
	ArmMoveToBelt
)
