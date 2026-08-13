package d2hero

import "testing"

// unknownSkillID is any ID not present in DevilSkills.
const unknownSkillID = -1

func TestCanLearnSkill(t *testing.T) {
	if !CanLearnSkill(SkillTraitDeFeu, 1) {
		t.Error("a level 1 hero should be able to learn Trait de feu (RequiredLevel 1)")
	}

	if CanLearnSkill(SkillTraitDeFeu, 0) {
		t.Error("a level 0 hero should not meet Trait de feu's RequiredLevel")
	}

	if CanLearnSkill(unknownSkillID, 99) {
		t.Error("an unknown skill ID should never be learnable, regardless of level")
	}
}

func TestSkillBaseSortDamageFallback(t *testing.T) {
	const fallback = 4

	if got := SkillBaseSortDamage(SkillTraitDeFeu, fallback); got != DevilSkills[SkillTraitDeFeu].BaseSortDamage {
		t.Errorf("expected Trait de feu's own base_sort, got %d", got)
	}

	if got := SkillBaseSortDamage(unknownSkillID, fallback); got != fallback {
		t.Errorf("expected fallback %d for an unknown skill, got %d", fallback, got)
	}
}

func TestSkillEclatDeGlaceIsDistinctFromTraitDeFeu(t *testing.T) {
	if SkillEclatDeGlace == SkillTraitDeFeu {
		t.Fatal("SkillEclatDeGlace must not collide with SkillTraitDeFeu")
	}

	fire, ice := DevilSkills[SkillTraitDeFeu], DevilSkills[SkillEclatDeGlace]

	if fire.Tree != ice.Tree {
		t.Error("expected both starting Élémentalisme skills to share a tree")
	}

	if fire.BaseSortDamage == ice.BaseSortDamage {
		t.Error("expected distinct base_sort values -- Éclat de glace trades damage for its slow effect")
	}
}

func TestSkillEclairEnChaineHasHigherTier(t *testing.T) {
	chain := DevilSkills[SkillEclairEnChaine]

	if chain.RequiredLevel <= DevilSkills[SkillTraitDeFeu].RequiredLevel {
		t.Error("expected Éclair en chaîne to require a higher level than the tier-1 starting skills")
	}

	if CanLearnSkill(SkillEclairEnChaine, chain.RequiredLevel-1) {
		t.Error("a hero below Éclair en chaîne's RequiredLevel should not be able to learn it")
	}

	if !CanLearnSkill(SkillEclairEnChaine, chain.RequiredLevel) {
		t.Error("a hero at exactly Éclair en chaîne's RequiredLevel should be able to learn it")
	}
}

func TestSkillNovaDeGivreSharesTierWithChainLightning(t *testing.T) {
	nova, chain := DevilSkills[SkillNovaDeGivre], DevilSkills[SkillEclairEnChaine]

	if nova.RequiredLevel != chain.RequiredLevel {
		t.Errorf("expected Nova de givre and Éclair en chaîne to share tier 2's RequiredLevel, got %d vs %d",
			nova.RequiredLevel, chain.RequiredLevel)
	}

	if nova.Tree != TreeElementalisme {
		t.Error("expected Nova de givre to belong to Élémentalisme")
	}
}

func TestSkillBouleDeFeuHasHigherTierThanTier2Spells(t *testing.T) {
	fireball := DevilSkills[SkillBouleDeFeu]

	if fireball.RequiredLevel <= DevilSkills[SkillNovaDeGivre].RequiredLevel {
		t.Error("expected Boule de feu to require a higher level than the tier-2 spells")
	}
}

func TestSkillMeteoreHasHigherTierAndDamageThanOrbeGlaciale(t *testing.T) {
	meteore := DevilSkills[SkillMeteore]

	if meteore.Tree != TreeElementalisme {
		t.Error("expected Météore to belong to Élémentalisme")
	}

	if meteore.RequiredLevel <= DevilSkills[SkillOrbeGlaciale].RequiredLevel {
		t.Error("expected Météore to require a higher level than the tier-4 Élémentalisme spells")
	}

	if meteore.BaseSortDamage <= DevilSkills[SkillOrbeGlaciale].BaseSortDamage {
		t.Error("expected Météore to deal more damage than the tier-4 Élémentalisme spells")
	}
}

func TestSkillApocalypseIsTheUltimate(t *testing.T) {
	apocalypse := DevilSkills[SkillApocalypse]

	for id, def := range DevilSkills {
		if id != SkillApocalypse && def.BaseSortDamage > apocalypse.BaseSortDamage {
			t.Errorf("expected Apocalypse (%d) to have the highest base_sort, but %q has %d",
				apocalypse.BaseSortDamage, def.Name, def.BaseSortDamage)
		}
	}

	if apocalypse.Tree != TreeElementalisme {
		t.Error("expected Apocalypse to belong to Élémentalisme")
	}

	if apocalypse.RequiredLevel < DevilSkills[SkillMeteore].RequiredLevel {
		t.Error("expected Apocalypse's tier to be at least Météore's (tier 5)")
	}
}

func TestSkillTempeteStatiqueSharesTierWithBouleDeFeu(t *testing.T) {
	storm, fireball := DevilSkills[SkillTempeteStatique], DevilSkills[SkillBouleDeFeu]

	if storm.RequiredLevel != fireball.RequiredLevel {
		t.Errorf("expected Tempête statique and Boule de feu to share tier 3's RequiredLevel, got %d vs %d",
			storm.RequiredLevel, fireball.RequiredLevel)
	}

	if storm.Tree != TreeElementalisme {
		t.Error("expected Tempête statique to belong to Élémentalisme")
	}
}

func TestSkillChampStatiqueIsArcane(t *testing.T) {
	if DevilSkills[SkillChampStatique].Tree != TreeArcane {
		t.Error("expected Champ statique to belong to Arcane")
	}
}

func TestSkillTelekinesieIsArcaneWithNoDamage(t *testing.T) {
	tk := DevilSkills[SkillTelekinesie]

	if tk.Tree != TreeArcane {
		t.Error("expected Télékinésie to belong to Arcane")
	}

	if tk.RequiredLevel != 1 {
		t.Errorf("expected Télékinésie to be a tier-1 skill, got RequiredLevel %d", tk.RequiredLevel)
	}

	if tk.BaseSortDamage != 0 {
		t.Errorf("expected Télékinésie to deal no direct base_sort damage, got %d", tk.BaseSortDamage)
	}
}

func TestSkillRalentissementSharesTierWithNovaDeGivre(t *testing.T) {
	slow, nova := DevilSkills[SkillRalentissement], DevilSkills[SkillNovaDeGivre]

	if slow.RequiredLevel != nova.RequiredLevel {
		t.Errorf("expected Ralentissement and Nova de givre to share tier 2's RequiredLevel, got %d vs %d",
			slow.RequiredLevel, nova.RequiredLevel)
	}

	if slow.Tree != TreeArcane {
		t.Error("expected Ralentissement to belong to Arcane")
	}

	if slow.BaseSortDamage != 0 {
		t.Errorf("expected Ralentissement to deal no direct base_sort damage, got %d", slow.BaseSortDamage)
	}
}

func TestSkillAmplificationSharesTierWithRalentissement(t *testing.T) {
	amp, slow := DevilSkills[SkillAmplification], DevilSkills[SkillRalentissement]

	if amp.RequiredLevel != slow.RequiredLevel {
		t.Errorf("expected Amplification and Ralentissement to share tier 2's RequiredLevel, got %d vs %d",
			amp.RequiredLevel, slow.RequiredLevel)
	}

	if amp.Tree != TreeArcane {
		t.Error("expected Amplification to belong to Arcane")
	}

	if amp.BaseSortDamage != 0 {
		t.Errorf("expected Amplification to deal no direct base_sort damage, got %d", amp.BaseSortDamage)
	}
}

func TestSkillTeleportationHasHigherTierThanAmplification(t *testing.T) {
	tp := DevilSkills[SkillTeleportation]

	if tp.RequiredLevel <= DevilSkills[SkillAmplification].RequiredLevel {
		t.Error("expected Téléportation to require a higher level than the tier-2 Arcane spells")
	}

	if tp.Tree != TreeArcane {
		t.Error("expected Téléportation to belong to Arcane")
	}

	if tp.BaseSortDamage != 0 {
		t.Errorf("expected Téléportation to deal no direct base_sort damage, got %d", tp.BaseSortDamage)
	}
}

func TestSkillOrbeGlacialeHasHighestTierAndLowerDamageThanBouleDeFeu(t *testing.T) {
	orb, fireball := DevilSkills[SkillOrbeGlaciale], DevilSkills[SkillBouleDeFeu]

	if orb.RequiredLevel <= fireball.RequiredLevel {
		t.Error("expected Orbe glaciale to require a higher level than the tier-3 Élémentalisme spells")
	}

	if orb.Tree != TreeElementalisme {
		t.Error("expected Orbe glaciale to belong to Élémentalisme")
	}

	if orb.BaseSortDamage >= fireball.BaseSortDamage {
		t.Error("expected Orbe glaciale to trade damage for its larger AoE footprint, like Nova de givre vs Trait de feu")
	}
}

func TestSkillPrisonDeGlaceHasHigherTierThanTeleportation(t *testing.T) {
	prison := DevilSkills[SkillPrisonDeGlace]

	if prison.RequiredLevel <= DevilSkills[SkillTeleportation].RequiredLevel {
		t.Error("expected Prison de glace to require a higher level than the tier-3 Arcane spells")
	}

	if prison.Tree != TreeArcane {
		t.Error("expected Prison de glace to belong to Arcane")
	}

	if prison.BaseSortDamage != 0 {
		t.Errorf("expected Prison de glace to deal no direct base_sort damage, got %d", prison.BaseSortDamage)
	}
}

func TestSkillVortexHasHigherTierThanPrisonDeGlace(t *testing.T) {
	vortex := DevilSkills[SkillVortex]

	if vortex.RequiredLevel <= DevilSkills[SkillPrisonDeGlace].RequiredLevel {
		t.Error("expected Vortex to require a higher level than the tier-4 Arcane spells")
	}

	if vortex.Tree != TreeArcane {
		t.Error("expected Vortex to belong to Arcane")
	}

	if vortex.BaseSortDamage != 0 {
		t.Errorf("expected Vortex to deal no direct base_sort damage, got %d", vortex.BaseSortDamage)
	}
}

func TestSkillDistorsionTemporelleIsHighestTierArcane(t *testing.T) {
	dt := DevilSkills[SkillDistorsionTemporelle]

	for id, def := range DevilSkills {
		if id != SkillDistorsionTemporelle && def.Tree == TreeArcane && def.RequiredLevel > dt.RequiredLevel {
			t.Errorf("expected Distorsion temporelle (%d) to be the highest-tier Arcane skill, but %q requires %d",
				dt.RequiredLevel, def.Name, def.RequiredLevel)
		}
	}

	if dt.BaseSortDamage != 0 {
		t.Errorf("expected Distorsion temporelle to deal no direct base_sort damage, got %d", dt.BaseSortDamage)
	}
}

func TestSkillBouclierDeManaIsDevilsFirstEsoterismeSkill(t *testing.T) {
	shield := DevilSkills[SkillBouclierDeMana]

	if shield.Tree != TreeEsoterisme {
		t.Errorf("expected Bouclier de mana to belong to Ésotérisme, got tree %v", shield.Tree)
	}

	if shield.RequiredLevel != 1 {
		t.Errorf("expected Bouclier de mana to be a tier-1 skill, got RequiredLevel %d", shield.RequiredLevel)
	}

	if shield.BaseSortDamage != 0 {
		t.Errorf("expected Bouclier de mana to deal no direct base_sort damage, got %d", shield.BaseSortDamage)
	}
}

func TestSkillArmureDeGlaceSharesTierWithArcaneTier2(t *testing.T) {
	shield := DevilSkills[SkillArmureDeGlace]

	if shield.Tree != TreeEsoterisme {
		t.Errorf("expected Armure de glace to belong to Ésotérisme, got tree %v", shield.Tree)
	}

	if shield.RequiredLevel <= DevilSkills[SkillBouclierDeMana].RequiredLevel {
		t.Error("expected Armure de glace to require a higher level than the tier-1 Ésotérisme skill")
	}

	if shield.BaseSortDamage != 0 {
		t.Errorf("expected Armure de glace to deal no direct base_sort damage, got %d", shield.BaseSortDamage)
	}
}

func TestSkillEveilDuNexusIsTheEsoterismeUltimate(t *testing.T) {
	nexus := DevilSkills[SkillEveilDuNexus]

	for id, def := range DevilSkills {
		if id != SkillEveilDuNexus && def.Tree == TreeEsoterisme && def.RequiredLevel > nexus.RequiredLevel {
			t.Errorf("expected Éveil du Nexus (%d) to be the highest-tier Ésotérisme skill, but %q requires %d",
				nexus.RequiredLevel, def.Name, def.RequiredLevel)
		}
	}

	if nexus.BaseSortDamage != 0 {
		t.Errorf("expected Éveil du Nexus to deal no direct base_sort damage, got %d", nexus.BaseSortDamage)
	}
}

func TestSkillAbsorptionEnergieIsAFreePassive(t *testing.T) {
	skill := DevilSkills[SkillAbsorptionEnergie]

	if skill.Tree != TreeEsoterisme {
		t.Errorf("expected Absorption d'énergie to belong to Ésotérisme, got tree %v", skill.Tree)
	}

	if skill.ManaCost != 0 {
		t.Errorf("expected Absorption d'énergie (a passive, never cast) to have ManaCost 0, got %d", skill.ManaCost)
	}

	if skill.BaseSortDamage != 0 {
		t.Errorf("expected Absorption d'énergie to deal no direct base_sort damage, got %d", skill.BaseSortDamage)
	}
}

func TestSkillTranscendanceHasHigherTierThanAbsorptionEnergie(t *testing.T) {
	skill := DevilSkills[SkillTranscendance]

	if skill.Tree != TreeEsoterisme {
		t.Errorf("expected Transcendance to belong to Ésotérisme, got tree %v", skill.Tree)
	}

	if skill.RequiredLevel <= DevilSkills[SkillAbsorptionEnergie].RequiredLevel {
		t.Error("expected Transcendance to require a higher level than the tier-3 Ésotérisme skill")
	}

	if skill.ManaCost != 0 {
		t.Errorf("expected Transcendance (a passive, never cast) to have ManaCost 0, got %d", skill.ManaCost)
	}
}

func TestSkillTempeteDeLamesHasHigherTierThanTranscendance(t *testing.T) {
	skill := DevilSkills[SkillTempeteDeLames]

	if skill.Tree != TreeEsoterisme {
		t.Errorf("expected Tempête de lames to belong to Ésotérisme, got tree %v", skill.Tree)
	}

	if skill.RequiredLevel <= DevilSkills[SkillTranscendance].RequiredLevel {
		t.Error("expected Tempête de lames to require a higher level than the tier-4 Ésotérisme skill")
	}
}

func TestSkillRuptureArcaneSharesTierWithTeleportation(t *testing.T) {
	rupture, tp := DevilSkills[SkillRuptureArcane], DevilSkills[SkillTeleportation]

	if rupture.RequiredLevel != tp.RequiredLevel {
		t.Errorf("expected Rupture arcane and Téléportation to share tier 3's RequiredLevel, got %d vs %d",
			rupture.RequiredLevel, tp.RequiredLevel)
	}

	if rupture.Tree != TreeArcane {
		t.Error("expected Rupture arcane to belong to Arcane")
	}

	if rupture.BaseSortDamage != 0 {
		t.Errorf("expected Rupture arcane to deal no direct base_sort damage, got %d", rupture.BaseSortDamage)
	}
}

func TestSkillManaCostFallback(t *testing.T) {
	const fallback = 2

	if got := SkillManaCost(SkillTraitDeFeu, fallback); got != DevilSkills[SkillTraitDeFeu].ManaCost {
		t.Errorf("expected Trait de feu's own mana cost, got %d", got)
	}

	if got := SkillManaCost(unknownSkillID, fallback); got != fallback {
		t.Errorf("expected fallback %d for an unknown skill, got %d", fallback, got)
	}
}
