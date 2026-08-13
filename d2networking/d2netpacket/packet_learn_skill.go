package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// LearnSkillRequestPacket is sent by a client, requesting that they learn
// the given skill (d2hero.HeroState.LearnSkill).
type LearnSkillRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	SkillID        int    `json:"skillId"`
}

// CreateLearnSkillRequestPacket returns a NetPacket which declares a
// LearnSkillRequestPacket for the given entity and skill.
func CreateLearnSkillRequestPacket(entityID string, skillID int) (NetPacket, error) {
	requestPacket := LearnSkillRequestPacket{
		SourceEntityID: entityID,
		SkillID:        skillID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.LearnSkillRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.LearnSkillRequest,
		PacketData: b,
	}, nil
}

// UnmarshalLearnSkillRequest unmarshals the given data to a
// LearnSkillRequestPacket struct.
func UnmarshalLearnSkillRequest(packet []byte) (LearnSkillRequestPacket, error) {
	var p LearnSkillRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SkillLearnedPacket carries a player's newly learned skill and their
// remaining SkillPoints. It is sent by the server whenever a
// LearnSkillRequestPacket successfully resolves.
type SkillLearnedPacket struct {
	PlayerID    string `json:"playerId"`
	SkillID     int    `json:"skillId"`
	SkillPoints int    `json:"skillPoints"`
}

// CreateSkillLearnedPacket returns a NetPacket which declares a
// SkillLearnedPacket for the given player, skill, and remaining points.
func CreateSkillLearnedPacket(playerID string, skillID, skillPoints int) (NetPacket, error) {
	learnedPacket := SkillLearnedPacket{
		PlayerID:    playerID,
		SkillID:     skillID,
		SkillPoints: skillPoints,
	}

	b, err := json.Marshal(learnedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SkillLearned}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SkillLearned,
		PacketData: b,
	}, nil
}

// UnmarshalSkillLearned unmarshals the given data to a SkillLearnedPacket
// struct.
func UnmarshalSkillLearned(packet []byte) (SkillLearnedPacket, error) {
	var p SkillLearnedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
