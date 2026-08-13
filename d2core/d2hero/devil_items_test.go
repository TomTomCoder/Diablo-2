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

func TestItemAllAttributesBonus(t *testing.T) {
	if got := ItemAllAttributesBonus(ItemAnneauDuDebut); got != 2 {
		t.Errorf("expected Anneau du Début's +2 to all attributes, got %d", got)
	}

	if got := ItemAllAttributesBonus("unknown-item-code"); got != 0 {
		t.Errorf("expected 0 for an unknown item code, got %d", got)
	}

	// sanity check: an item with no AllAttributesBonus set (e.g. the
	// staff, which only grants Energy/Fire%) shouldn't report one either.
	if got := ItemAllAttributesBonus(ItemBatonApprenti); got != 0 {
		t.Errorf("expected Bâton de l'Apprenti to have no all-attributes bonus, got %d", got)
	}
}
