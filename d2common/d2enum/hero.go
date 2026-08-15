package d2enum

import "log"

//go:generate stringer -linecomment -type Hero
//go:generate string2enum -samepkg -linecomment -type Hero

// Hero is used for different types of hero's
type Hero int

// Heroes
const (
	HeroNone        Hero = iota //
	HeroBarbarian               // Barbarian
	HeroNecromancer             // Necromancer
	HeroPaladin                 // Paladin
	HeroAssassin                // Assassin
	HeroSorceress               // Sorceress
	HeroAmazon                  // Amazon
	HeroDruid                   // Druid
	// HeroDevil is Devil's own hero token (ROADMAP.md Phase 6) --
	// previously Devil reused one of the seven Diablo II class tokens
	// above purely for cosmetic resource lookups (skill icon sheets,
	// inventory panel layout...); every one of those lookups now has a
	// HeroDevil case pointing at the same generic/placeholder resources
	// Devil's own skills already fall back to (Charclass "", see
	// d2hero.NewDevilHeroSkill), so this is additive -- nothing that used
	// to work for the seven real classes changes.
	HeroDevil // Devil
)

// GetToken returns a 2 letter token
func (h Hero) GetToken() string {
	switch h {
	case HeroBarbarian:
		return "BA"
	case HeroNecromancer:
		return "NE"
	case HeroPaladin:
		return "PA"
	case HeroAssassin:
		return "AI"
	case HeroSorceress:
		return "SO"
	case HeroAmazon:
		return "AM"
	case HeroDruid:
		return "DZ"
	case HeroDevil:
		return "DE"
	default:
		log.Fatalf("Unknown hero token: %d", h)
	}

	return ""
}

// CompositeToken returns the token to use specifically for loading a
// hero's in-game character composite (d2asset.AssetManager.LoadComposite,
// see d2mapentity.MapEntityFactory.NewPlayer) -- identical to GetToken for
// the seven real Diablo II classes, but for HeroDevil returns
// HeroSorceress's own real token ("SO") instead of "DE".
//
// This is a deliberate, documented placeholder skin, not a mistake:
// LoadComposite panics on any error, and no .COF/.DC6 files exist for the
// "DE" token anywhere (ROADMAP.md Phase 6, no real Devil art yet) -- so
// composite loading is the one call site GetToken's real "DE" identity
// cannot safely reach. Sorceress is reused because it's already Devil's
// established "closest thematic fit" everywhere else a missing asset
// needs a stand-in (a caster class, same as the inventory panel layout in
// GameControls and the skill icon fallback sheet in NewDevilHeroSkill).
// GetToken/GetToken3 themselves are untouched and still return Devil's
// real "DE"/"DEV" identity -- everything that isn't asset loading (saves,
// class-specific game logic, display) keeps seeing the real class.
func (h Hero) CompositeToken() string {
	if h == HeroDevil {
		return HeroSorceress.GetToken()
	}

	return h.GetToken()
}

// GetToken3 returns a 3 letter token
func (h Hero) GetToken3() string {
	switch h {
	case HeroBarbarian:
		return "BAR"
	case HeroNecromancer:
		return "NEC"
	case HeroPaladin:
		return "PAL"
	case HeroAssassin:
		return "ASS"
	case HeroSorceress:
		return "SOR"
	case HeroAmazon:
		return "AMA"
	case HeroDruid:
		return "DRU"
	case HeroDevil:
		return "DEV"
	default:
		log.Fatalf("Unknown hero token: %d", h)
	}

	return ""
}
