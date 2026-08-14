package d2player

import (
	"strconv"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

const (
	skillLabelXOffset = 49
	skillLabelYOffset = -4

	skillIconXOff  = 346
	skillIconYOff  = 59
	skillIconDistX = 69
	skillIconDistY = 68
)

func newSkillIcon(ui *d2ui.UIManager,
	baseSprite *d2ui.Sprite,
	l d2util.LogLevel,
	skill *d2hero.HeroSkill) *skillIcon {
	base := d2ui.NewBaseWidget(ui)
	label := ui.NewLabel(d2resource.Font16, d2resource.PaletteSky)

	x := skillIconXOff + skill.SkillColumn*skillIconDistX
	y := skillIconYOff + skill.SkillRow*skillIconDistY

	res := &skillIcon{
		BaseWidget: base,
		sprite:     baseSprite,
		skill:      skill,
		lvlLabel:   label,
		enabled:    true,
	}

	res.Logger = d2util.NewLogger()
	res.Logger.SetLevel(l)
	res.Logger.SetPrefix(logPrefix)

	res.SetPosition(x, y)

	// ponytail: BaseWidget defaults to a zero-size hit box, so without this
	// Contains() (and therefore every click) would always miss -- the icon
	// would render but never be clickable. iw/ih fall back to 0 on error
	// (e.g. IconCel out of range for a placeholder skill), same failure
	// mode as renderSprite already tolerates below.
	if iw, ih, err := baseSprite.GetFrameSize(skill.IconCel); err == nil {
		res.SetSize(iw, ih)
	}

	return res
}

type skillIcon struct {
	*d2ui.BaseWidget
	lvlLabel   *d2ui.Label
	sprite     *d2ui.Sprite
	skill      *d2hero.HeroSkill
	enabled    bool
	pressed    bool
	onActivate func()

	*d2util.Logger
}

// static check: WidgetGroup.AddWidget only registers a widget as
// clickable via a runtime type assertion (w.(d2ui.ClickableWidget)) --
// nothing else would catch a mismatch here, same reasoning as
// widget_group.go's own `var _ Widget = &WidgetGroup{}`.
var _ d2ui.ClickableWidget = &skillIcon{}

// The four methods below plus OnActivated/Activate satisfy
// d2ui.ClickableWidget. Implementing it is what makes a skillIcon
// clickable at all: UIManager.addClickable (called from AddWidget for
// anything satisfying this interface) is the only place mouse-up events
// turn into an Activate() call -- see d2core/d2ui/ui_manager.go.

// GetEnabled returns whether the icon currently accepts clicks.
func (si *skillIcon) GetEnabled() bool { return si.enabled }

// SetEnabled sets whether the icon currently accepts clicks.
func (si *skillIcon) SetEnabled(enabled bool) { si.enabled = enabled }

// GetPressed returns whether the icon is currently held down.
func (si *skillIcon) GetPressed() bool { return si.pressed }

// SetPressed sets whether the icon is currently held down.
func (si *skillIcon) SetPressed(pressed bool) { si.pressed = pressed }

// OnActivated registers the callback run when the icon is clicked.
func (si *skillIcon) OnActivated(callback func()) { si.onActivate = callback }

// Activate runs the click callback, if one was registered.
func (si *skillIcon) Activate() {
	if si.onActivate != nil {
		si.onActivate()
	}
}

func (si *skillIcon) SetVisible(visible bool) {
	si.BaseWidget.SetVisible(visible)
	si.lvlLabel.SetVisible(visible)
}

func (si *skillIcon) renderSprite(target d2interface.Surface) {
	x, y := si.GetPosition()

	if err := si.sprite.SetCurrentFrame(si.skill.IconCel); err != nil {
		si.Errorf("Cannot set Frame %e", err)
		return
	}

	if si.skill.SkillPoints == 0 {
		target.PushSaturation(skillIconGreySat)
		defer target.Pop()

		target.PushBrightness(skillIconGreyBright)
		defer target.Pop()
	}

	si.sprite.SetPosition(x, y)
	si.sprite.Render(target)
}

func (si *skillIcon) renderSpriteLabel(target d2interface.Surface) {
	if si.skill.SkillPoints == 0 {
		return
	}

	x, y := si.GetPosition()
	si.lvlLabel.SetText(strconv.Itoa(si.skill.SkillPoints))
	si.lvlLabel.SetPosition(x+skillLabelXOffset, y+skillLabelYOffset)
	si.lvlLabel.Render(target)
}

func (si *skillIcon) Render(target d2interface.Surface) {
	si.renderSprite(target)
	si.renderSpriteLabel(target)
}

func (si *skillIcon) Advance(elapsed float64) error {
	return nil
}
