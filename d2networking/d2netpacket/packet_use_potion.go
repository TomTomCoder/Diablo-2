package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// UsePotionRequestPacket is sent by a client, requesting that the potion in
// BeltIndex of their own belt (d2hero.HeroState.Belt) be consumed.
type UsePotionRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	BeltIndex      int    `json:"beltIndex"`
}

// CreateUsePotionRequestPacket returns a NetPacket which declares a
// UsePotionRequestPacket for the given entity and belt slot.
func CreateUsePotionRequestPacket(entityID string, beltIndex int) (NetPacket, error) {
	requestPacket := UsePotionRequestPacket{
		SourceEntityID: entityID,
		BeltIndex:      beltIndex,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.UsePotionRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.UsePotionRequest,
		PacketData: b,
	}, nil
}

// UnmarshalUsePotionRequest unmarshals the given data to a
// UsePotionRequestPacket struct.
func UnmarshalUsePotionRequest(packet []byte) (UsePotionRequestPacket, error) {
	var p UsePotionRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// PotionUsedPacket carries a player's new Mana total after consuming a
// potion. It is sent by the server whenever a UsePotionRequestPacket
// successfully resolves.
type PotionUsedPacket struct {
	PlayerID string `json:"playerId"`
	Mana     int    `json:"mana"`
}

// CreatePotionUsedPacket returns a NetPacket which declares a
// PotionUsedPacket with the given player's new Mana total.
func CreatePotionUsedPacket(playerID string, mana int) (NetPacket, error) {
	usedPacket := PotionUsedPacket{
		PlayerID: playerID,
		Mana:     mana,
	}

	b, err := json.Marshal(usedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.PotionUsed}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.PotionUsed,
		PacketData: b,
	}, nil
}

// UnmarshalPotionUsed unmarshals the given data to a PotionUsedPacket
// struct.
func UnmarshalPotionUsed(packet []byte) (PotionUsedPacket, error) {
	var p PotionUsedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
