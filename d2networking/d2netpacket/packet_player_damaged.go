package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// PlayerDamagedPacket contains the result of a hit resolved against a
// player. It is sent by the server so clients can update the player's HP.
//
// Mana included too, not just HP -- correction (août 2026): Bouclier de
// mana (HeroStatsState.ApplyDamageWithManaShield) drains Mana instead of
// Health while active, as a direct side effect of the same hit this packet
// already reports, but the client never learned about it since only HP was
// ever carried.
type PlayerDamagedPacket struct {
	PlayerID string `json:"playerId"`
	HP       int    `json:"hp"`
	Mana     int    `json:"mana"`
	Died     bool   `json:"died"`
}

// CreatePlayerDamagedPacket returns a NetPacket which declares a
// PlayerDamagedPacket with the given hit result and current Mana.
func CreatePlayerDamagedPacket(playerID string, hp, mana int, died bool) (NetPacket, error) {
	playerDamagedPacket := PlayerDamagedPacket{
		PlayerID: playerID,
		HP:       hp,
		Mana:     mana,
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
