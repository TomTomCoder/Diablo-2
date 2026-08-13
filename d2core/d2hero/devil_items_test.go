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

func TestItemEnergyBonus(t *testing.T) {
	if got := ItemEnergyBonus(ItemPendentifArcane); got != 3 {
		t.Errorf("expected Pendentif Arcane's +3 Energy, got %d", got)
	}

	if got := ItemEnergyBonus("unknown-item-code"); got != 0 {
		t.Errorf("expected 0 for an unknown item code, got %d", got)
	}
}
