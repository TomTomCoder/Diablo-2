package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// NPCMovedPacket carries an NPC's new position after a server-authoritative
// instant displacement -- Télékinésie's Knockback and Vortex's Pull, same
// shape as NPCHitPacket but for position instead of HP. It is sent by the
// server whenever such a displacement happens, so clients move the local
// copy of the entity to match instead of it silently staying put.
type NPCMovedPacket struct {
	EntityID string  `json:"entityId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// CreateNPCMovedPacket returns a NetPacket which declares an NPCMovedPacket
// for the given entity and its new position.
func CreateNPCMovedPacket(entityID string, x, y float64) (NetPacket, error) {
	movedPacket := NPCMovedPacket{
		EntityID: entityID,
		X:        x,
		Y:        y,
	}

	b, err := json.Marshal(movedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.NPCMoved}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.NPCMoved,
		PacketData: b,
	}, nil
}

// UnmarshalNPCMoved unmarshals the given data to an NPCMovedPacket struct.
func UnmarshalNPCMoved(packet []byte) (NPCMovedPacket, error) {
	var p NPCMovedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
