package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// PlayerTeleportedPacket carries a player's new, instantly-set position. It
// is sent by the server after Téléportation, and is meant to snap the
// client's copy of the player directly there -- unlike MovePlayerPacket,
// which paths a smooth walk between two points.
type PlayerTeleportedPacket struct {
	PlayerID string  `json:"playerId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// CreatePlayerTeleportedPacket returns a NetPacket which declares a
// PlayerTeleportedPacket with the given player's new position.
func CreatePlayerTeleportedPacket(playerID string, x, y float64) (NetPacket, error) {
	teleportedPacket := PlayerTeleportedPacket{
		PlayerID: playerID,
		X:        x,
		Y:        y,
	}

	b, err := json.Marshal(teleportedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.PlayerTeleported}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.PlayerTeleported,
		PacketData: b,
	}, nil
}

// UnmarshalPlayerTeleported unmarshals the given data to a
// PlayerTeleportedPacket struct.
func UnmarshalPlayerTeleported(packet []byte) (PlayerTeleportedPacket, error) {
	var p PlayerTeleportedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
