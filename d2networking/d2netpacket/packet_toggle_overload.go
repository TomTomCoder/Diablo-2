package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// ToggleOverloadRequestPacket is sent by a client, requesting that
// Surcharge (Overload) be toggled on/off (HeroStatsState.OverloadActive
// directly, via GameServer.resolveToggleOverload) -- unlike Bouclier de
// mana/Armure de glace's own toggles, Overload isn't one of the 30 tree
// skills (§6 lists it as a separate generic mechanic), so it has no
// SkillID to carry through a CastPacket the way those two do.
type ToggleOverloadRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
}

// CreateToggleOverloadRequestPacket returns a NetPacket which declares a
// ToggleOverloadRequestPacket for the given entity.
func CreateToggleOverloadRequestPacket(entityID string) (NetPacket, error) {
	requestPacket := ToggleOverloadRequestPacket{
		SourceEntityID: entityID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ToggleOverloadRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ToggleOverloadRequest,
		PacketData: b,
	}, nil
}

// UnmarshalToggleOverloadRequest unmarshals the given data to a
// ToggleOverloadRequestPacket struct.
func UnmarshalToggleOverloadRequest(packet []byte) (ToggleOverloadRequestPacket, error) {
	var p ToggleOverloadRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
