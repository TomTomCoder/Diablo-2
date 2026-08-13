package d2netpacket

import (
	"encoding/json"
	"time"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// Player status effect names carried by PlayerStatusEffectPacket.Effect --
// matching the d2hero.HeroStatsState field/method each names.
const (
	PlayerStatusManaShield    = "manaShield"    // ManaShieldActive
	PlayerStatusArmureDeGlace = "armureDeGlace" // ArmureDeGlaceActive
	PlayerStatusMagicImmune   = "magicImmune"   // MagicImmuneUntil (via Until, Active unused)
)

// PlayerStatusEffectPacket carries a change to one of the caster's own
// status effects (Bouclier de mana/Armure de glace's toggles, Éveil du
// Nexus's timed immunity). Sent by the server whenever one of those three
// mutates HeroStatsState -- previously none of them broadcast anything, so
// even in solo play (which still round-trips through a local
// client/server, per ROADMAP.md) the client's own copy of its player never
// learned about them, unlike every other Devil mechanic this session.
type PlayerStatusEffectPacket struct {
	PlayerID string    `json:"playerId"`
	Effect   string    `json:"effect"`
	Active   bool      `json:"active"` // toggles (ManaShield/ArmureDeGlace): the new on/off state
	Until    time.Time `json:"until"`  // timed effects (MagicImmune): when it expires
}

// CreatePlayerStatusEffectPacket returns a NetPacket which declares a
// PlayerStatusEffectPacket for the given player, effect, toggle state, and
// expiry.
func CreatePlayerStatusEffectPacket(playerID, effect string, active bool, until time.Time) (NetPacket, error) {
	statusPacket := PlayerStatusEffectPacket{
		PlayerID: playerID,
		Effect:   effect,
		Active:   active,
		Until:    until,
	}

	b, err := json.Marshal(statusPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.PlayerStatusEffect}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.PlayerStatusEffect,
		PacketData: b,
	}, nil
}

// UnmarshalPlayerStatusEffect unmarshals the given data to a
// PlayerStatusEffectPacket struct.
func UnmarshalPlayerStatusEffect(packet []byte) (PlayerStatusEffectPacket, error) {
	var p PlayerStatusEffectPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
