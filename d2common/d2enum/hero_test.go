package d2enum

import "testing"

// TestHeroDevilTokensDoNotFatal guards against the real failure mode
// GetToken/GetToken3's own default case uses: log.Fatalf, which terminates
// the whole process (not a recoverable panic) for any Hero value missing
// a case there. Go tests can't assert on a fatal exit, so this just calls
// both and requires a real, non-empty, collision-free token back --
// proof the HeroDevil case was actually added, not proof-by-absence-of-
// crash (which a test process dying wouldn't even report cleanly).
func TestHeroDevilTokensDoNotFatal(t *testing.T) {
	existingTokens := map[string]bool{
		"BA": true, "NE": true, "PA": true, "AI": true, "SO": true, "AM": true, "DZ": true,
	}
	existingTokens3 := map[string]bool{
		"BAR": true, "NEC": true, "PAL": true, "ASS": true, "SOR": true, "AMA": true, "DRU": true,
	}

	token := HeroDevil.GetToken()
	if token == "" || existingTokens[token] {
		t.Errorf("expected a real, non-colliding 2-letter token for HeroDevil, got %q", token)
	}

	token3 := HeroDevil.GetToken3()
	if token3 == "" || existingTokens3[token3] {
		t.Errorf("expected a real, non-colliding 3-letter token for HeroDevil, got %q", token3)
	}
}

// TestHeroDevilCompositeTokenUsesSorceressPlaceholder is a regression
// test for the fix that unblocked an actual playable Devil character:
// CompositeToken must return Sorceress's real token for HeroDevil (no
// .COF/.DC6 files exist for GetToken's real "DE"), while GetToken itself
// stays untouched.
func TestHeroDevilCompositeTokenUsesSorceressPlaceholder(t *testing.T) {
	if got, want := HeroDevil.CompositeToken(), HeroSorceress.GetToken(); got != want {
		t.Errorf("expected HeroDevil.CompositeToken() to be Sorceress's own token %q, got %q", want, got)
	}

	if HeroDevil.GetToken() == HeroDevil.CompositeToken() {
		t.Error("expected CompositeToken to differ from GetToken's real (unusable) \"DE\" identity")
	}
}

// TestCompositeTokenMatchesGetTokenForRealClasses is a regression test:
// CompositeToken must be a pure pass-through to GetToken for every real
// Diablo II class -- only HeroDevil gets a placeholder substitution.
func TestCompositeTokenMatchesGetTokenForRealClasses(t *testing.T) {
	realClasses := []Hero{
		HeroBarbarian, HeroNecromancer, HeroPaladin, HeroAssassin, HeroSorceress, HeroAmazon, HeroDruid,
	}

	for _, hero := range realClasses {
		if got, want := hero.CompositeToken(), hero.GetToken(); got != want {
			t.Errorf("expected CompositeToken() == GetToken() for %v, got %q vs %q", hero, got, want)
		}
	}
}

func TestHeroDevilStringAndFromStringRoundTrip(t *testing.T) {
	if HeroDevil.String() != "Devil" {
		t.Errorf(`expected HeroDevil.String() == "Devil", got %q`, HeroDevil.String())
	}

	if HeroFromString("Devil") != HeroDevil {
		t.Errorf("expected HeroFromString(\"Devil\") == HeroDevil, got %v", HeroFromString("Devil"))
	}
}
