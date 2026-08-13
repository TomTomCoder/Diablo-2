package d2hero

import "testing"

func TestItemFireDamagePercent(t *testing.T) {
	if got := ItemFireDamagePercent(ItemBatonApprenti); got != 10 {
		t.Errorf("expected Bâton de l'Apprenti's +10%% Fire modifier, got %d", got)
	}

	if got := ItemFireDamagePercent("unknown-item-code"); got != 0 {
		t.Errorf("expected 0 for an unknown item code, got %d", got)
	}
}
