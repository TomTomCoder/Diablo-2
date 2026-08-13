package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// PlayerDamagedPacket contains the result of a hit resolved against a
// player. It is sent by the server so clients can update the player's HP.
type PlayerDamagedPacket struct {
	PlayerID string `json:"playerId"`
	HP       int    `json:"hp"`
	Died     bool   `json:"died"`
}

// CreatePlayerDamagedPacket returns a NetPacket which declares a
// PlayerDamagedPacket with the given hit result.
func CreatePlayerDamagedPacket(playerID string, hp int, died bool) (NetPacket, error) {
	playerDamagedPacket := PlayerDamagedPacket{
		PlayerID: playerID,
		HP:       hp,
		Died:     died,
	}

	b, err := json.Marshal(playerDamagedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.PlayerDamaged}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.PlayerDamaged,
		PacketData: b,
	}, nil
}

// UnmarshalPlayerDamaged unmarshals the given data to a PlayerDamagedPacket struct
func UnmarshalPlayerDamaged(packet []byte) (PlayerDamagedPacket, error) {
	var p PlayerDamagedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
