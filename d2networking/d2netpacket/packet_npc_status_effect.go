package d2netpacket

import (
	"encoding/json"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// NPC status effect names carried by NPCStatusEffectPacket.Effect --
// matching the d2mapentity.NPC method each names (ApplySlow/
// ApplyImmobilize/ApplyAmplification/ApplyResistanceStrip).
const (
	NPCStatusSlowed             = "slowed"
	NPCStatusImmobilized        = "immobilized"
	NPCStatusAmplified          = "amplified"
	NPCStatusResistanceStripped = "resistanceStripped"
)

// NPCStatusEffectPacket carries a timed status effect applied to an NPC
// (Éclat de glace/Ralentissement/Distorsion temporelle's slow, Prison de
// glace's immobilize, Amplification's damage amplification, Rupture
// arcane's resistance strip). One generic packet covers all four rather
// than one packet type per effect, since they're all the same shape: an
// entity ID and an expiry.
type NPCStatusEffectPacket struct {
	EntityID string    `json:"entityId"`
	Effect   string    `json:"effect"`
	Until    time.Time `json:"until"`
}

// CreateNPCStatusEffectPacket returns a NetPacket which declares an
// NPCStatusEffectPacket for the given entity, effect, and expiry.
func CreateNPCStatusEffectPacket(entityID, effect string, until time.Time) (NetPacket, error) {
	statusPacket := NPCStatusEffectPacket{
		EntityID: entityID,
		Effect:   effect,
		Until:    until,
	}

	b, err := json.Marshal(statusPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.NPCStatusEffect}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.NPCStatusEffect,
		PacketData: b,
	}, nil
}

// UnmarshalNPCStatusEffect unmarshals the given data to an
// NPCStatusEffectPacket struct.
func UnmarshalNPCStatusEffect(packet []byte) (NPCStatusEffectPacket, error) {
	var p NPCStatusEffectPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
