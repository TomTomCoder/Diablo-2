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
)
