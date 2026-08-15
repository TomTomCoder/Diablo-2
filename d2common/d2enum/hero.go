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
