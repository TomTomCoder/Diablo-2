package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// RespecSkillsRequestPacket is sent by a client, requesting that every
// skill they've learned be forgotten and refunded
// (d2hero.HeroState.RespecSkills -- "Respec partiel").
type RespecSkillsRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
}

// CreateRespecSkillsRequestPacket returns a NetPacket which declares a
// RespecSkillsRequestPacket for the given entity.
func CreateRespecSkillsRequestPacket(entityID string) (NetPacket, error) {
	requestPacket := RespecSkillsRequestPacket{SourceEntityID: entityID}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.RespecSkillsRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.RespecSkillsRequest,
		PacketData: b,
	}, nil
}

// UnmarshalRespecSkillsRequest unmarshals the given data to a
// RespecSkillsRequestPacket struct.
func UnmarshalRespecSkillsRequest(packet []byte) (RespecSkillsRequestPacket, error) {
	var p RespecSkillsRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SkillsRespecedPacket carries a player's new SkillPoints total after every
// learned skill was forgotten and refunded. It is sent by the server
// whenever a RespecSkillsRequestPacket successfully resolves.
//
// The attribute/pool fields below exist because HeroState.RespecSkills
// ("Respec partiel", devil_game_design_reference.md §10: "tous les points
// de compétences et d'attributs") also refunds every attribute point ever
// spent (HeroStatsState.RespecAllAttributePoints) -- correction (août
// 2026): the original packet only ever carried SkillPoints, so the
// client's own copy of the caster's attributes/StatsPoints/health/mana
// pools never learned about the other half of the same action.
type SkillsRespecedPacket struct {
	PlayerID    string `json:"playerId"`
	SkillPoints int    `json:"skillPoints"`
	StatsPoints int    `json:"statsPoints"`
	Strength    int    `json:"strength"`
	Energy      int    `json:"energy"`
	Dexterity   int    `json:"dexterity"`
	Vitality    int    `json:"vitality"`
	MaxHealth   int    `json:"maxHealth"`
	Health      int    `json:"health"`
	MaxMana     int    `json:"maxMana"`
	Mana        int    `json:"mana"`
}

// CreateSkillsRespecedPacket returns a NetPacket which declares a
// SkillsRespecedPacket for the given player, carrying their post-respec
// SkillPoints and full attribute/pool state.
func CreateSkillsRespecedPacket(
	playerID string, skillPoints, statsPoints, strength, energy, dexterity, vitality, maxHealth, health, maxMana, mana int,
) (NetPacket, error) {
	respecedPacket := SkillsRespecedPacket{
		PlayerID:    playerID,
		SkillPoints: skillPoints,
		StatsPoints: statsPoints,
		Strength:    strength,
		Energy:      energy,
		Dexterity:   dexterity,
		Vitality:    vitality,
		MaxHealth:   maxHealth,
		Health:      health,
		MaxMana:     maxMana,
		Mana:        mana,
	}

	b, err := json.Marshal(respecedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SkillsRespeced}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SkillsRespeced,
		PacketData: b,
	}, nil
}

// UnmarshalSkillsRespeced unmarshals the given data to a
// SkillsRespecedPacket struct.
func UnmarshalSkillsRespeced(packet []byte) (SkillsRespecedPacket, error) {
	var p SkillsRespecedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// RespecSingleSkillRequestPacket is sent by a client, requesting that one
// specific learned skill be forgotten and refunded
// (d2hero.HeroState.RespecSingleSkill -- "Glyphe d'oubli").
type RespecSingleSkillRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	SkillID        int    `json:"skillId"`
}

// CreateRespecSingleSkillRequestPacket returns a NetPacket which declares a
// RespecSingleSkillRequestPacket for the given entity and skill.
func CreateRespecSingleSkillRequestPacket(entityID string, skillID int) (NetPacket, error) {
	requestPacket := RespecSingleSkillRequestPacket{
		SourceEntityID: entityID,
		SkillID:        skillID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.RespecSingleSkillRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.RespecSingleSkillRequest,
		PacketData: b,
	}, nil
}

// UnmarshalRespecSingleSkillRequest unmarshals the given data to a
// RespecSingleSkillRequestPacket struct.
func UnmarshalRespecSingleSkillRequest(packet []byte) (RespecSingleSkillRequestPacket, error) {
	var p RespecSingleSkillRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SingleSkillRespecedPacket carries the one skill a player forgot and their
// new SkillPoints total. It is sent by the server whenever a
// RespecSingleSkillRequestPacket successfully resolves.
type SingleSkillRespecedPacket struct {
	PlayerID    string `json:"playerId"`
	SkillID     int    `json:"skillId"`
	SkillPoints int    `json:"skillPoints"`
}

// CreateSingleSkillRespecedPacket returns a NetPacket which declares a
// SingleSkillRespecedPacket for the given player, forgotten skill, and new
// SkillPoints total.
func CreateSingleSkillRespecedPacket(playerID string, skillID, skillPoints int) (NetPacket, error) {
	respecedPacket := SingleSkillRespecedPacket{
		PlayerID:    playerID,
		SkillID:     skillID,
		SkillPoints: skillPoints,
	}

	b, err := json.Marshal(respecedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SingleSkillRespeced}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SingleSkillRespeced,
		PacketData: b,
	}, nil
}

// UnmarshalSingleSkillRespeced unmarshals the given data to a
// SingleSkillRespecedPacket struct.
func UnmarshalSingleSkillRespeced(packet []byte) (SingleSkillRespecedPacket, error) {
	var p SingleSkillRespecedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
