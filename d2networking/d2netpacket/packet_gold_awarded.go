package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// GoldAwardedPacket carries a player's new gold total. It is sent by the
// server whenever gold is awarded (e.g. a monster's death drop).
type GoldAwardedPacket struct {
	PlayerID string `json:"playerId"`
	Gold     int    `json:"gold"`
}

// CreateGoldAwardedPacket returns a NetPacket which declares a
// GoldAwardedPacket with the given player's new gold total.
func CreateGoldAwardedPacket(playerID string, gold int) (NetPacket, error) {
	goldAwardedPacket := GoldAwardedPacket{
		PlayerID: playerID,
		Gold:     gold,
	}

	b, err := json.Marshal(goldAwardedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.GoldAwarded}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.GoldAwarded,
		PacketData: b,
	}, nil
}

// UnmarshalGoldAwarded unmarshals the given data to a GoldAwardedPacket struct
func UnmarshalGoldAwarded(packet []byte) (GoldAwardedPacket, error) {
	var p GoldAwardedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
