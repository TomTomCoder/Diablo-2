package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// EquipSkillRequestPacket is sent by a client, requesting that an
// already-learned skill be assigned to one of their two active-skill slots
// (d2hero.HeroState.EquipSkill). Slot is a d2hero.SkillSlot value, carried
// as a plain int the same way d2hero.Attribute already crosses the network
// in SpendAttributePointRequestPacket.
type EquipSkillRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	Slot           int    `json:"slot"`
	SkillID        int    `json:"skillId"`
}

// CreateEquipSkillRequestPacket returns a NetPacket which declares an
// EquipSkillRequestPacket for the given entity, slot, and skill.
func CreateEquipSkillRequestPacket(entityID string, slot, skillID int) (NetPacket, error) {
	requestPacket := EquipSkillRequestPacket{
		SourceEntityID: entityID,
		Slot:           slot,
		SkillID:        skillID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.EquipSkillRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.EquipSkillRequest,
		PacketData: b,
	}, nil
}

// UnmarshalEquipSkillRequest unmarshals the given data to an
// EquipSkillRequestPacket struct.
func UnmarshalEquipSkillRequest(packet []byte) (EquipSkillRequestPacket, error) {
	var p EquipSkillRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SkillEquippedPacket carries a player's newly assigned skill slot. It is
// sent by the server whenever an EquipSkillRequestPacket successfully
// resolves.
type SkillEquippedPacket struct {
	PlayerID string `json:"playerId"`
	Slot     int    `json:"slot"`
	SkillID  int    `json:"skillId"`
}

// CreateSkillEquippedPacket returns a NetPacket which declares a
// SkillEquippedPacket for the given player, slot, and skill.
func CreateSkillEquippedPacket(playerID string, slot, skillID int) (NetPacket, error) {
	equippedPacket := SkillEquippedPacket{
		PlayerID: playerID,
		Slot:     slot,
		SkillID:  skillID,
	}

	b, err := json.Marshal(equippedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SkillEquipped}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SkillEquipped,
		PacketData: b,
	}, nil
}

// UnmarshalSkillEquipped unmarshals the given data to a SkillEquippedPacket
// struct.
func UnmarshalSkillEquipped(packet []byte) (SkillEquippedPacket, error) {
	var p SkillEquippedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
