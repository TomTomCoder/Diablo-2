package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// NPCHitPacket contains the result of a hit resolved against an NPC. It is
// sent by the server so clients can update (or remove, if Died) the entity.
type NPCHitPacket struct {
	EntityID string `json:"entityId"`
	HP       int    `json:"hp"`
	Died     bool   `json:"died"`
}

// CreateNPCHitPacket returns a NetPacket which declares an NPCHitPacket with
// the given hit result.
func CreateNPCHitPacket(entityID string, hp int, died bool) (NetPacket, error) {
	npcHitPacket := NPCHitPacket{
		EntityID: entityID,
		HP:       hp,
		Died:     died,
	}

	b, err := json.Marshal(npcHitPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.NPCHit}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.NPCHit,
		PacketData: b,
	}, nil
}

// UnmarshalNPCHit unmarshals the given data to an NPCHitPacket struct
func UnmarshalNPCHit(packet []byte) (NPCHitPacket, error) {
	var p NPCHitPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
