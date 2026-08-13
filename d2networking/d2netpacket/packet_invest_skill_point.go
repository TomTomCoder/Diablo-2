package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// InvestSkillPointRequestPacket is sent by a client, requesting that they
// invest another skill point into a skill they already know
// (d2hero.HeroState.InvestSkillPoint).
type InvestSkillPointRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	SkillID        int    `json:"skillId"`
}

// CreateInvestSkillPointRequestPacket returns a NetPacket which declares an
// InvestSkillPointRequestPacket for the given entity and skill.
func CreateInvestSkillPointRequestPacket(entityID string, skillID int) (NetPacket, error) {
	requestPacket := InvestSkillPointRequestPacket{
		SourceEntityID: entityID,
		SkillID:        skillID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.InvestSkillPointRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.InvestSkillPointRequest,
		PacketData: b,
	}, nil
}

// UnmarshalInvestSkillPointRequest unmarshals the given data to an
// InvestSkillPointRequestPacket struct.
func UnmarshalInvestSkillPointRequest(packet []byte) (InvestSkillPointRequestPacket, error) {
	var p InvestSkillPointRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SkillPointInvestedPacket carries a skill's new total invested points
// (HeroSkill.SkillPoints) and the caster's remaining hero-level
// SkillPoints. It is sent by the server whenever an
// InvestSkillPointRequestPacket successfully resolves.
type SkillPointInvestedPacket struct {
	PlayerID       string `json:"playerId"`
	SkillID        int    `json:"skillId"`
	InvestedPoints int    `json:"investedPoints"` // the skill's new HeroSkill.SkillPoints total
	SkillPoints    int    `json:"skillPoints"`    // the caster's remaining hero-level skill points
}

// CreateSkillPointInvestedPacket returns a NetPacket which declares a
// SkillPointInvestedPacket for the given player, skill, its new invested
// total, and the caster's remaining SkillPoints.
func CreateSkillPointInvestedPacket(playerID string, skillID, investedPoints, skillPoints int) (NetPacket, error) {
	investedPacket := SkillPointInvestedPacket{
		PlayerID:       playerID,
		SkillID:        skillID,
		InvestedPoints: investedPoints,
		SkillPoints:    skillPoints,
	}

	b, err := json.Marshal(investedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SkillPointInvested}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SkillPointInvested,
		PacketData: b,
	}, nil
}

// UnmarshalSkillPointInvested unmarshals the given data to a
// SkillPointInvestedPacket struct.
func UnmarshalSkillPointInvested(packet []byte) (SkillPointInvestedPacket, error) {
	var p SkillPointInvestedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
