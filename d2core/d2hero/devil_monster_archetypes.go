package d2hero

import "time"

// MonsterArchetypeDef overrides the flat per-monster combat placeholders
// (aggro radius, attack range, damage, attack cooldown) for a specific
// monster type, keyed by its stable monster key
// (d2records.MonStatRecord.Key, "Id" in monstats.txt -- see
// d2mapentity.NPC.MonsterKey). Devil's own data model, same
// struct+registry+lookup-with-fallback shape as DevilSkillDef/DevilItemDef.
type MonsterArchetypeDef struct {
	Key                 string
	AggroRadiusSubtiles float64
	AttackRangeSubtiles float64
	AttackCooldown      time.Duration
	AttackDamage        int
}

// MonsterArchetypes is the registry of Devil's own monster archetypes,
// keyed by monster key.
//
// ponytail: empty for now -- no Devil monster data exists yet, same wall as
// the ally-summon Ésotérisme skills and Phase 6's visual assets (a real
// d2mapentity.NPC needs a real d2records.MonStatRecord, which needs real
// monster data). The lookup/fallback functions below are already wired
// into d2networking/d2server/game_server.go's monster AI
// (advanceMonsterAI/tryMonsterAttack), so per-monster-type tuning works the
// moment real monster data exists -- it won't need another round of
// server-side changes, just registry entries here.
var MonsterArchetypes = map[string]*MonsterArchetypeDef{}

// MonsterAggroRadiusSubtiles returns key's aggro radius, or fallback if
// key isn't registered.
func MonsterAggroRadiusSubtiles(key string, fallback float64) float64 {
	if def, ok := MonsterArchetypes[key]; ok {
		return def.AggroRadiusSubtiles
	}

	return fallback
}

// MonsterAttackRangeSubtiles returns key's attack range, or fallback if
// key isn't registered.
func MonsterAttackRangeSubtiles(key string, fallback float64) float64 {
	if def, ok := MonsterArchetypes[key]; ok {
		return def.AttackRangeSubtiles
	}

	return fallback
}

// MonsterAttackCooldown returns key's attack cooldown, or fallback if key
// isn't registered.
func MonsterAttackCooldown(key string, fallback time.Duration) time.Duration {
	if def, ok := MonsterArchetypes[key]; ok {
		return def.AttackCooldown
	}

	return fallback
}

// MonsterAttackDamage returns key's attack damage, or fallback if key
// isn't registered.
func MonsterAttackDamage(key string, fallback int) int {
	if def, ok := MonsterArchetypes[key]; ok {
		return def.AttackDamage
	}

	return fallback
}
