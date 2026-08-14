package d2hero

import "github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"

// HydrateSkills will load the SkillRecord & SkillDescriptionRecord from the asset manager, using the skill ID.
// This is done to avoid serializing the whole record data of HeroSkill to a game save or network packets.
// We cant do this while unmarshalling because there is no reference to the asset manager.
//
// Correction: this previously assumed every skill ID lives in D2's own
// skills.txt (asset.Records.Skill.Details), which is never true for
// Devil's own namespaced skill IDs (SkillTraitDeFeu and friends,
// deliberately outside D2's real ID ranges) -- the map lookup missed,
// returning nil, and the very next line dereferenced that nil
// *d2records.SkillRecord's Skilldesc field, panicking. This fired for
// every player on their very first AddPlayer packet, since even the
// starting Trait de feu is a Devil skill ID (select_hero_class.go).
func HydrateSkills(skills map[int]*HeroSkill, asset *d2asset.AssetManager) {
	for skillID, skill := range skills {
		if _, isDevilSkill := DevilSkills[skillID]; isDevilSkill {
			// Devil's own skills aren't in D2's skills.txt --
			// NewDevilHeroSkill already built everything a HeroSkill needs
			// at creation time, nothing to hydrate here.
			continue
		}

		skill.SkillRecord = asset.Records.Skill.Details[skillID]
		if skill.SkillRecord == nil {
			continue
		}

		skill.SkillDescriptionRecord = asset.Records.Skill.Descriptions[skill.SkillRecord.Skilldesc]
	}
}
