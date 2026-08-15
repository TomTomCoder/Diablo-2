package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// ToggleMarqueArdenteRequestPacket is sent by a client, requesting that
// Marque ardente be toggled on/off (HeroStatsState.MarqueArdenteActive
// directly, via GameServer.resolveToggleMarqueArdente) -- same shape as
// ToggleOverloadRequestPacket, see its own doc comment: Marque ardente is
// §6's other generic mechanic, also with no SkillID to carry through a
// CastPacket.
type ToggleMarqueArdenteRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
}

// CreateToggleMarqueArdenteRequestPacket returns a NetPacket which
// declares a ToggleMarqueArdenteRequestPacket for the given entity.
func CreateToggleMarqueArdenteRequestPacket(entityID string) (NetPacket, error) {
	requestPacket := ToggleMarqueArdenteRequestPacket{
		SourceEntityID: entityID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ToggleMarqueArdenteRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ToggleMarqueArdenteRequest,
		PacketData: b,
	}, nil
}

// UnmarshalToggleMarqueArdenteRequest unmarshals the given data to a
// ToggleMarqueArdenteRequestPacket struct.
func UnmarshalToggleMarqueArdenteRequest(packet []byte) (ToggleMarqueArdenteRequestPacket, error) {
	var p ToggleMarqueArdenteRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
