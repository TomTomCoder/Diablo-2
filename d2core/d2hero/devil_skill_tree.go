package d2hero

import "sort"

// SkillTree identifies which of Devil's three skill trees a skill belongs
// to (devil_game_design_reference.md §7).
type SkillTree int

// Devil's three skill trees.
const (
	TreeElementalisme SkillTree = iota // dégâts directs
	TreeArcane                         // contrôle/amplification
	TreeEsoterisme                     // défense/survie/invocations
)

// DevilSkillDef is Devil's own skill data model. Deliberately separate from
// Diablo 2's skills.txt/SkillRecord schema: different mana-cost shape, and
// per-point synergies (§7 "Règles des synergies") modeled as standalone
// registries next to the specific mechanic each one boosts (see
// TraitDeFeuSynergyTargets/BouclierDeManaSynergyReductionPercent in
// devil_skills.go) rather than a generic field on this struct -- D2's own
// synergies vary too much in shape (damage bonus, duration, reduction...) for
// one shared field to fit them all. See ROADMAP.md Phase 2 for why reusing
// D2's columns wasn't a good fit.
type DevilSkillDef struct {
	ID             int
	Name           string
	Tree           SkillTree
	RequiredLevel  int // "palier" -- the character level needed to spend a point here
	BaseSortDamage int
	ManaCost       int

	// DealsFireDamage marks a skill whose damage is (at least partly) Fire,
	// per its own design description (Trait de feu/Boule de feu/Météore's
	// "dégâts feu", Apocalypse's "feu" among its three elements). Gates
	// equipment Fire modifiers (d2hero.ItemFireDamagePercent) in
	// resolveAttackDamage -- a bool rather than a full multi-value element
	// field since Fire is the only element any Devil item currently grants
	// a bonus to; add Cold/Lightning fields the same way if/when an item
	// ever does.
	DealsFireDamage bool

	// Passive marks a skill that's never cast -- an always-on bonus once
	// learned (Absorption d'énergie/Transcendance/Maîtrise
	// élémentaire/Résonance magique/Régénération accélérée, each already
	// documented "passive -- never cast" at their own definition below).
	// Explicit field rather than inferring it from ManaCost == 0, so
	// nothing downstream has to guess: d2game/d2player's skill-select
	// popup (skill_select_panel.go) needs to exclude these from the
	// equippable list the same way D2's own SkillRecord.Passive does.
	Passive bool
}

// SkillEclatDeGlace is Devil's own skill ID for "Éclat de glace"
// (Élémentalisme, devil_game_design_reference.md §7): "Projectile froid,
// ralentit la cible". Its slow effect lands on d2mapentity.NPC via
// ApplySlow -- see resolveMeleeHit in game_server.go.
//
// ponytail: same namespacing reasoning as SkillTraitDeFeu -- numbered right
// after it, still nowhere near a real D2 skills.txt ID range.
const SkillEclatDeGlace = SkillTraitDeFeu + 1

// SkillEclairEnChaine is Devil's own skill ID for "Éclair en chaîne"
// (Élémentalisme, devil_game_design_reference.md §7): "Foudre qui rebondit
// sur 3 cibles". Its multi-target hit is resolved by resolveChainHit in
// game_server.go -- see skillEclairEnChaineTargets/eclairEnChaineChainRadius.
const SkillEclairEnChaine = SkillEclatDeGlace + 1

// SkillNovaDeGivre is Devil's own skill ID for "Nova de givre"
// (Élémentalisme, devil_game_design_reference.md §7): "Explosion de froid en
// zone autour du Mage". Unlike the other Élémentalisme spells, it's centered
// on the caster rather than a targeted/nearest NPC -- see resolveNovaHit in
// game_server.go.
const SkillNovaDeGivre = SkillEclairEnChaine + 1

// SkillBouleDeFeu is Devil's own skill ID for "Boule de feu" (Élémentalisme,
// devil_game_design_reference.md §7): "Projectile AoE, dégâts feu élevés".
// Like Nova de givre it's an AoE hit (see resolveAoeHit in game_server.go),
// but centered on the cast's targeted position rather than the caster.
const SkillBouleDeFeu = SkillNovaDeGivre + 1

// SkillTempeteStatique is Devil's own skill ID for "Tempête statique"
// (Élémentalisme, devil_game_design_reference.md §7): "Invoque une zone
// d'éclair persistante".
//
// ponytail: resolved as a single instant AoE hit at the targeted position
// (same resolveAoeHit path as Boule de feu) rather than an actual
// persistent, re-ticking zone -- no zone/duration entity exists yet.
// Upgrade path: a per-cast expiry tracked the same way lastCastAt/
// lastMonsterAttackAt are, re-running resolveAoeHit on that position every
// aiTickInterval until it lapses.
const SkillTempeteStatique = SkillBouleDeFeu + 1

// SkillOrbeGlaciale is Devil's own skill ID for "Orbe glaciale"
// (Élémentalisme, devil_game_design_reference.md §7): "Projectile lent,
// explose en large AoE de froid". Same AoE-at-target-position shape as
// Boule de feu (see resolveAoeHit in game_server.go), just a larger radius
// and a lower base_sort -- the value here is area coverage, not raw damage.
const SkillOrbeGlaciale = SkillTempeteStatique + 1

// SkillChampStatique is Devil's own skill ID for "Champ statique" (Arcane,
// devil_game_design_reference.md §7): "Réduit la vie de toutes les entités à
// l'écran d'un % fixe". Devil's first Arcane skill -- percent-of-current-HP
// damage rather than base_sort scaling, so BaseSortDamage is unused (0) for
// it. See resolveChampStatiqueHit in game_server.go.
const SkillChampStatique = SkillOrbeGlaciale + 1

// SkillTelekinesie is Devil's own skill ID for "Télékinésie" (Arcane,
// devil_game_design_reference.md §7): "Repousse les entités, interaction
// avec les objets à distance". Only the knockback half is modeled --
// see resolveTelekinesieHit in game_server.go and d2mapentity.NPC.Knockback.
// The "interaction avec les objets à distance" half needs the item/ground-
// interaction system, which doesn't exist yet (ROADMAP.md Phase 5).
const SkillTelekinesie = SkillChampStatique + 1

// SkillRalentissement is Devil's own skill ID for "Ralentissement" (Arcane,
// devil_game_design_reference.md §7): "Zone qui réduit la vitesse des
// entités de 50%". Deals no damage (BaseSortDamage 0) -- its whole value is
// the slow, applied to every killable NPC in the targeted area via the same
// d2mapentity.NPC.ApplySlow used by Éclat de glace. See
// resolveRalentissementHit in game_server.go.
const SkillRalentissement = SkillTelekinesie + 1

// SkillAmplification is Devil's own skill ID for "Amplification" (Arcane,
// devil_game_design_reference.md §7): "Augmente les dégâts magiques reçus
// par la cible". A single-target debuff -- see the inline dispatch in
// GameServer.resolveMeleeHit (game_server.go) and
// d2mapentity.NPC.ApplyAmplification/IsAmplified. Deals
// no direct damage itself (BaseSortDamage 0).
const SkillAmplification = SkillRalentissement + 1

// SkillTeleportation is Devil's own skill ID for "Téléportation" (Arcane,
// devil_game_design_reference.md §7): "Déplacement instantané vers la
// position visée". Moves the caster themselves, not an NPC -- see
// resolveTeleportationHit in game_server.go and the PlayerTeleported
// packet. Deals no damage (BaseSortDamage 0).
const SkillTeleportation = SkillAmplification + 1

// SkillPrisonDeGlace is Devil's own skill ID for "Prison de glace" (Arcane,
// devil_game_design_reference.md §7): "Immobilise un groupe d'entités".
// Unlike Ralentissement's partial slow, this forces every killable NPC in
// the targeted area to a dead stop -- see resolvePrisonDeGlaceHit in
// game_server.go and d2mapentity.NPC.ApplyImmobilize/IsImmobilized. Deals
// no direct damage (BaseSortDamage 0).
const SkillPrisonDeGlace = SkillTeleportation + 1

// SkillVortex is Devil's own skill ID for "Vortex" (Arcane,
// devil_game_design_reference.md §7): "Aspire toutes les entités proches
// vers un point". Knockback's inverse -- see resolveVortexHit in
// game_server.go and d2mapentity.NPC.Pull. Deals no direct damage
// (BaseSortDamage 0).
const SkillVortex = SkillPrisonDeGlace + 1

// SkillMeteore is Devil's own skill ID for "Météore" (Élémentalisme,
// devil_game_design_reference.md §7): "Frappe retardée sur une zone,
// dégâts feu massifs". Same AoE-at-target-position shape as Boule de feu
// (see resolveAoeHit in game_server.go) -- highest base_sort of any
// Élémentalisme spell so far.
//
// ponytail: resolved as an instant AoE hit rather than an actual delayed
// strike, same simplification (and same reasoning) as Tempête statique's
// missing "persistante" zone -- no delayed-effect/timer entity exists yet.
const SkillMeteore = SkillVortex + 1

// SkillDistorsionTemporelle is Devil's own skill ID for "Distorsion
// temporelle" (Arcane, devil_game_design_reference.md §7): "Ralentit toutes
// les entités à l'écran pendant 5 secondes". Devil's highest-tier Arcane
// skill. Unlike Ralentissement's own placeholder duration, the design gives
// an exact number here -- see distorsionTemporelleSlowDuration in
// game_server.go. Deals no damage (BaseSortDamage 0).
const SkillDistorsionTemporelle = SkillMeteore + 1

// SkillApocalypse is Devil's own skill ID for "Apocalypse" (Élémentalisme,
// devil_game_design_reference.md §7): "Pluie d'éclairs + feu + froid sur
// toute la zone visible". Élémentalisme's ultimate (highest tier, highest
// base_sort) -- hits every killable NPC on the map via the normal
// Energy-scaled resolveAttackDamage/applyHit path (unlike Champ statique's
// percent-of-HP damage), same "toute la zone visible" modeling choice as
// Champ statique/Distorsion temporelle (every NPC on the map, not a real
// per-client screen/viewport query). See resolveApocalypseHit in
// game_server.go.
const SkillApocalypse = SkillDistorsionTemporelle + 1

// SkillBouclierDeMana is Devil's own skill ID for "Bouclier de mana"
// (Ésotérisme, devil_game_design_reference.md §7): "Absorbe les dégâts avec
// la réserve de mana". Devil's first Ésotérisme skill. Casting it toggles
// HeroStatsState.ManaShieldActive -- see resolveBouclierDeManaHit in
// game_server.go. The absorption itself already existed from an earlier
// phase (HeroStatsState.ApplyDamageWithManaShield, wired into
// tryMonsterAttack); this finally gives the player a real way to turn it on
// and off, rather than it being set directly for tests only. Deals no
// direct damage (BaseSortDamage 0).
const SkillBouclierDeMana = SkillApocalypse + 1

// SkillArmureDeGlace is Devil's own skill ID for "Armure de glace"
// (Ésotérisme, devil_game_design_reference.md §7): "Réduit les dégâts
// reçus et ralentit les attaquants au contact". A toggle, same shape as
// Bouclier de mana -- see resolveArmureDeGlaceHit in game_server.go and
// HeroStatsState.ArmureDeGlaceActive. Deals no direct damage (BaseSortDamage
// 0).
const SkillArmureDeGlace = SkillBouclierDeMana + 1

// SkillEveilDuNexus is Devil's own skill ID for "Éveil du Nexus"
// (Ésotérisme, devil_game_design_reference.md §7): "Ultime défensif :
// immunité magique pendant 8 secondes, soigne le Mage". Ésotérisme's
// ultimate -- see resolveEveilDuNexusHit in game_server.go and
// HeroStatsState.ApplyMagicImmunity/IsMagicImmune/Heal. Deals no direct
// damage (BaseSortDamage 0).
const SkillEveilDuNexus = SkillArmureDeGlace + 1

// SkillAbsorptionEnergie is Devil's own skill ID for "Absorption d'énergie"
// (Ésotérisme, devil_game_design_reference.md §7): "Chaque entité tuée
// restaure un % de mana". Passif, but -- unlike Maîtrise
// élémentaire/Résonance magique -- doesn't need per-point scaling to be
// meaningful: it's a flat restore, on or off, so the existing
// learned/not-learned model (HeroState.Skills) is enough. See
// GameServer.restoreManaOnKill. Deals no direct damage (BaseSortDamage 0).
const SkillAbsorptionEnergie = SkillEveilDuNexus + 1

// SkillTranscendance is Devil's own skill ID for "Transcendance"
// (Ésotérisme, devil_game_design_reference.md §7): "Au lieu de mourir, le
// Mage se régénère une fois par zone (longue recharge)". A true passive
// like Absorption d'énergie -- learned/not-learned is enough, no per-point
// scaling needed. See GameServer.tryTranscend. Deals no direct damage
// (BaseSortDamage 0), never cast (ManaCost 0).
const SkillTranscendance = SkillAbsorptionEnergie + 1

// SkillTempeteDeLames is Devil's own skill ID for "Tempête de lames"
// (Ésotérisme, devil_game_design_reference.md §7): "Invoque des lames de
// mana orbitant autour du Mage". Same self-centered-AoE shape as Nova de
// givre (see resolveSelfCenteredAoeHit in game_server.go) -- the
// "orbiting/periodic" part isn't modeled, same simplification as Tempête
// statique's missing "persistante" zone.
const SkillTempeteDeLames = SkillTranscendance + 1

// SkillRuptureArcane is Devil's own skill ID for "Rupture arcane" (Arcane,
// devil_game_design_reference.md §7): "Projectile qui supprime les
// résistances d'une cible". A single-target debuff, same tier as
// Téléportation -- see the inline dispatch in GameServer.resolveMeleeHit
// (game_server.go) and
// d2mapentity.NPC.ApplyResistanceStrip/IsResistanceStripped. Deals no direct
// damage itself (BaseSortDamage 0).
const SkillRuptureArcane = SkillTempeteDeLames + 1

// SkillMaitriseElementaire is Devil's own skill ID for "Maîtrise
// élémentaire" (Élémentalisme, devil_game_design_reference.md §7): "Augmente
// tous les dégâts élémentaires (+% par point)". Devil's first skill that
// takes more than one invested point -- see HeroState.InvestSkillPoint and
// MaitriseElementaireDamagePercent. Never cast (BaseSortDamage/ManaCost 0).
const SkillMaitriseElementaire = SkillRuptureArcane + 1

// SkillResonanceMagique is Devil's own skill ID for "Résonance magique"
// (Arcane, devil_game_design_reference.md §7): "Synergie : chaque sort
// lancé augmente les dégâts du suivant (+% temporaire)". Devil's second
// multi-point-investment skill (see HeroState.InvestSkillPoint), same tier
// as Prison de glace. Unlike Maîtrise élémentaire's permanent bonus, this
// one is a temporary window armed by every cast and consumed by the next
// eligible one -- see HeroStatsState.ApplyResonanceMagiqueBonus/
// ConsumeResonanceMagiqueBonus and ResonanceMagiqueDamagePercent. §7
// "Synergies": it amplifies "tous les sorts actifs des arbres I et II"
// (Élémentalisme and Arcane only, not Ésotérisme). Never cast
// (BaseSortDamage/ManaCost 0).
const SkillResonanceMagique = SkillMaitriseElementaire + 1

// SkillRegenerationAcceleree is Devil's own skill ID for "Régénération
// accélérée" (Ésotérisme, devil_game_design_reference.md §7): "Augmente la
// vitesse de régénération du mana". Same tier as Bouclier de mana, Devil's
// third multi-point-investment skill -- see
// GameServer.manaRegenPerSecond's call site in game_server.go and
// RegenerationAccelereePercent. Never cast (BaseSortDamage/ManaCost 0).
const SkillRegenerationAcceleree = SkillResonanceMagique + 1

// DevilSkills is the registry of Devil's own skill data, keyed by ID.
//
// 27 of the design's 30 skills are here (ROADMAP.md Phase 2) -- the 3
// missing ones (Familier, Double ésotérique, Golem arcane) are genuine ally
// summons, blocked on real monster data + sprites that don't exist yet
// (see MapEntityFactory.NewNPC's requirements), not a data-model gap.
//
// Re-verified (août 2026), not just re-asserted: NewNPC's sprite half
// (LoadComposite's COF/DC6/AnimData) is exactly what the placeholder
// asset pipeline built this session (d2core/d2asset/placeholdergen)
// already proves out end-to-end -- that part of this wall is gone. What
// still blocks these 3 skills is the *data* half: NewNPC also requires a
// full d2records.MonStatRecord (HP range, speed, equipment options,
// weapon class...), and none of these three creatures has any specified
// stats anywhere in the design docs. Inventing HP/damage/speed numbers
// for them would cross the same "no invented balance numbers" line the
// rest of this codebase already holds -- the wall moved, it didn't fall.
//
// nolint:gochecknoglobals // a read-only registry, not mutable shared state
// -- flagged now that golangci-lint actually runs (août 2026).
var DevilSkills = map[int]*DevilSkillDef{
	SkillTraitDeFeu: {
		ID:              SkillTraitDeFeu,
		Name:            "Trait de feu",
		Tree:            TreeElementalisme,
		RequiredLevel:   1,
		BaseSortDamage:  6,
		ManaCost:        3,
		DealsFireDamage: true,
	},
	SkillEclatDeGlace: {
		ID:             SkillEclatDeGlace,
		Name:           "Éclat de glace",
		Tree:           TreeElementalisme,
		RequiredLevel:  1,
		BaseSortDamage: 4, // lower base damage than Trait de feu -- its value is the slow, not raw damage
		ManaCost:       4,
	},
	SkillEclairEnChaine: {
		ID:             SkillEclairEnChaine,
		Name:           "Éclair en chaîne",
		Tree:           TreeElementalisme,
		RequiredLevel:  6, // tier 2 of Élémentalisme, per the design's level table
		BaseSortDamage: 5,
		ManaCost:       6,
	},
	SkillNovaDeGivre: {
		ID:             SkillNovaDeGivre,
		Name:           "Nova de givre",
		Tree:           TreeElementalisme,
		RequiredLevel:  6, // tier 2, alongside Éclair en chaîne
		BaseSortDamage: 4, // same cold-family trade-off as Éclat de glace: lower damage, an AoE footprint instead
		ManaCost:       8, // costliest tier-2 spell -- it hits every killable NPC in range, not just one
	},
	SkillBouleDeFeu: {
		ID:              SkillBouleDeFeu,
		Name:            "Boule de feu",
		Tree:            TreeElementalisme,
		RequiredLevel:   12, // tier 3, per the design's level table
		BaseSortDamage:  10, // "dégâts feu élevés" -- highest base_sort of any Élémentalisme spell so far
		ManaCost:        10,
		DealsFireDamage: true,
	},
	SkillTempeteStatique: {
		ID:             SkillTempeteStatique,
		Name:           "Tempête statique",
		Tree:           TreeElementalisme,
		RequiredLevel:  12, // tier 3, alongside Boule de feu
		BaseSortDamage: 6,  // lightning-family: between Éclair en chaîne and Boule de feu
		ManaCost:       9,
	},
	SkillOrbeGlaciale: {
		ID:             SkillOrbeGlaciale,
		Name:           "Orbe glaciale",
		Tree:           TreeElementalisme,
		RequiredLevel:  18, // tier 4, per the design's level table
		BaseSortDamage: 5,  // lower than the other tier-3/4 spells -- its value is the AoE footprint, not raw damage
		ManaCost:       11,
	},
	SkillChampStatique: {
		ID:            SkillChampStatique,
		Name:          "Champ statique",
		Tree:          TreeArcane,
		RequiredLevel: 1,
		ManaCost:      12, // costliest tier-1 spell -- it hits every killable NPC on the map, not just one
	},
	SkillTelekinesie: {
		ID:            SkillTelekinesie,
		Name:          "Télékinésie",
		Tree:          TreeArcane,
		RequiredLevel: 1,
		ManaCost:      5, // cheap utility/control spell -- no direct damage
	},
	SkillRalentissement: {
		ID:            SkillRalentissement,
		Name:          "Ralentissement",
		Tree:          TreeArcane,
		RequiredLevel: 6, // tier 2, per the design's level table
		ManaCost:      7,
	},
	SkillAmplification: {
		ID:            SkillAmplification,
		Name:          "Amplification",
		Tree:          TreeArcane,
		RequiredLevel: 6, // tier 2, alongside Ralentissement
		ManaCost:      6,
	},
	SkillTeleportation: {
		ID:            SkillTeleportation,
		Name:          "Téléportation",
		Tree:          TreeArcane,
		RequiredLevel: 12, // tier 3, per the design's level table
		ManaCost:      8,  // utility spell -- costs more than the tier-2 debuffs, less than a tier-3 AoE
	},
	SkillPrisonDeGlace: {
		ID:            SkillPrisonDeGlace,
		Name:          "Prison de glace",
		Tree:          TreeArcane,
		RequiredLevel: 18, // tier 4, per the design's level table
		ManaCost:      13, // costliest Arcane spell so far -- full immobilization over an area
	},
	SkillVortex: {
		ID:            SkillVortex,
		Name:          "Vortex",
		Tree:          TreeArcane,
		RequiredLevel: 24, // tier 5, per the design's level table
		ManaCost:      12,
	},
	SkillMeteore: {
		ID:              SkillMeteore,
		Name:            "Météore",
		Tree:            TreeElementalisme,
		RequiredLevel:   24, // tier 5, per the design's level table
		BaseSortDamage:  14, // "dégâts feu massifs" -- highest base_sort of any Devil spell so far
		ManaCost:        14,
		DealsFireDamage: true,
	},
	SkillDistorsionTemporelle: {
		ID:            SkillDistorsionTemporelle,
		Name:          "Distorsion temporelle",
		Tree:          TreeArcane,
		RequiredLevel: 30, // tier 6, per the design's level table -- Devil's highest-tier Arcane skill
		ManaCost:      15, // costliest Arcane spell so far
	},
	SkillApocalypse: {
		ID:             SkillApocalypse,
		Name:           "Apocalypse",
		Tree:           TreeElementalisme,
		RequiredLevel:  30, // tier 6, per the design's level table -- Devil's highest-tier Élémentalisme skill
		BaseSortDamage: 16, // the ultimate -- highest base_sort of any Devil spell
		ManaCost:       18, // costliest spell in the game so far
		// "Pluie d'éclairs + feu + froid" -- one of its three elements is
		// Fire, so it still qualifies for a Fire weapon bonus.
		DealsFireDamage: true,
	},
	SkillBouclierDeMana: {
		ID:            SkillBouclierDeMana,
		Name:          "Bouclier de mana",
		Tree:          TreeEsoterisme,
		RequiredLevel: 1,
		ManaCost:      4, // cheap toggle -- the real cost is the mana drained per hit absorbed, not the cast itself
	},
	SkillArmureDeGlace: {
		ID:            SkillArmureDeGlace,
		Name:          "Armure de glace",
		Tree:          TreeEsoterisme,
		RequiredLevel: 6, // tier 2, per the design's level table
		ManaCost:      6,
	},
	SkillEveilDuNexus: {
		ID:            SkillEveilDuNexus,
		Name:          "Éveil du Nexus",
		Tree:          TreeEsoterisme,
		RequiredLevel: 30, // tier 6, per the design's level table -- Devil's highest-tier Ésotérisme skill
		ManaCost:      16, // costly defensive ultimate
	},
	SkillAbsorptionEnergie: {
		ID:            SkillAbsorptionEnergie,
		Name:          "Absorption d'énergie",
		Tree:          TreeEsoterisme,
		RequiredLevel: 12, // tier 3, per the design's level table
		ManaCost:      0,  // passive -- never cast, always on once learned
		Passive:       true,
	},
	SkillTranscendance: {
		ID:            SkillTranscendance,
		Name:          "Transcendance",
		Tree:          TreeEsoterisme,
		RequiredLevel: 18, // tier 4, per the design's level table
		ManaCost:      0,  // passive -- never cast, always on once learned
		Passive:       true,
	},
	SkillTempeteDeLames: {
		ID:             SkillTempeteDeLames,
		Name:           "Tempête de lames",
		Tree:           TreeEsoterisme,
		RequiredLevel:  24, // tier 5, per the design's level table
		BaseSortDamage: 8,
		ManaCost:       12,
	},
	SkillRuptureArcane: {
		ID:            SkillRuptureArcane,
		Name:          "Rupture arcane",
		Tree:          TreeArcane,
		RequiredLevel: 12, // tier 3, alongside Téléportation
		ManaCost:      7,  // single-target debuff -- between Amplification's 6 (tier 2) and Téléportation's 8 (tier 3 utility)
	},
	SkillMaitriseElementaire: {
		ID:            SkillMaitriseElementaire,
		Name:          "Maîtrise élémentaire",
		Tree:          TreeElementalisme,
		RequiredLevel: 18, // tier 4, alongside Orbe glaciale
		ManaCost:      0,  // passive -- never cast, always on once learned
		Passive:       true,
	},
	SkillResonanceMagique: {
		ID:            SkillResonanceMagique,
		Name:          "Résonance magique",
		Tree:          TreeArcane,
		RequiredLevel: 18, // tier 4, alongside Prison de glace
		ManaCost:      0,  // passive -- never cast, always on once learned
		Passive:       true,
	},
	SkillRegenerationAcceleree: {
		ID:            SkillRegenerationAcceleree,
		Name:          "Régénération accélérée",
		Tree:          TreeEsoterisme,
		RequiredLevel: 1, // tier 1, alongside Bouclier de mana
		ManaCost:      0, // passive -- never cast, always on once learned
		Passive:       true,
	},
}

// CanLearnSkill reports whether a hero of the given level may spend a
// point in skillID, per its RequiredLevel ("palier" gating,
// devil_game_design_reference.md §7). False for an unknown skill ID.
func CanLearnSkill(skillID, heroLevel int) bool {
	def, ok := DevilSkills[skillID]
	if !ok {
		return false
	}

	return heroLevel >= def.RequiredLevel
}

// SkillGridPosition returns the skill tree UI page/column/row for skillID,
// derived deterministically from its Tree/RequiredLevel/ID rather than a
// per-skill layout the design never specifies (it only specifies which
// tree and tier a skill belongs to, devil_game_design_reference.md §7) --
// this exists purely so d2game/d2player's skill tree panel has *some*
// distinct, stable position to render each skill at, instead of every
// skill defaulting to (0,0) and rendering on top of each other.
//
// page is 1-indexed (1/2/3, matching d2game/d2player/skilltree.go's
// firstTab/secondTab/thirdTab+1 convention) -- Élémentalisme/Arcane/
// Ésotérisme, in SkillTree's own iota order. Within a tree, skills are
// ordered by RequiredLevel (the design's own tier progression) then ID as
// a stable tiebreak, laid out 2 per row so ~10 skills per tree stays a
// reasonable height rather than one very tall single column. Returns all
// zeroes for an unknown skillID.
func SkillGridPosition(skillID int) (page, column, row int) {
	def, ok := DevilSkills[skillID]
	if !ok {
		return 0, 0, 0
	}

	const columns = 2

	sameTree := make([]int, 0, len(DevilSkills))

	for id, other := range DevilSkills {
		if other.Tree == def.Tree {
			sameTree = append(sameTree, id)
		}
	}

	sort.Slice(sameTree, func(i, j int) bool {
		a, b := DevilSkills[sameTree[i]], DevilSkills[sameTree[j]]
		if a.RequiredLevel != b.RequiredLevel {
			return a.RequiredLevel < b.RequiredLevel
		}

		return sameTree[i] < sameTree[j]
	})

	for index, id := range sameTree {
		if id == skillID {
			return int(def.Tree) + 1, index % columns, index / columns
		}
	}

	return int(def.Tree) + 1, 0, 0
}

// SkillBaseSortDamage returns skillID's base_sort damage, or fallback if
// it's not in DevilSkills.
func SkillBaseSortDamage(skillID, fallback int) int {
	if def, ok := DevilSkills[skillID]; ok {
		return def.BaseSortDamage
	}

	return fallback
}

// SkillManaCost returns skillID's mana cost, or fallback if it's not in
// DevilSkills.
func SkillManaCost(skillID, fallback int) int {
	if def, ok := DevilSkills[skillID]; ok {
		return def.ManaCost
	}

	return fallback
}
