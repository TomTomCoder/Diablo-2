package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// SpendAttributePointRequestPacket is sent by a client, requesting that
// they spend one of their level-up attribute points on the given attribute
// (d2hero.HeroStatsState.SpendAttributePoint). Attribute is the int value
// of a d2hero.Attribute (Strength/Energy/Dexterity/Vitality).
type SpendAttributePointRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	Attribute      int    `json:"attribute"`
}

// CreateSpendAttributePointRequestPacket returns a NetPacket which declares
// a SpendAttributePointRequestPacket for the given entity and attribute.
func CreateSpendAttributePointRequestPacket(entityID string, attribute int) (NetPacket, error) {
	requestPacket := SpendAttributePointRequestPacket{
		SourceEntityID: entityID,
		Attribute:      attribute,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SpendAttributePointRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SpendAttributePointRequest,
		PacketData: b,
	}, nil
}

// UnmarshalSpendAttributePointRequest unmarshals the given data to a
// SpendAttributePointRequestPacket struct.
func UnmarshalSpendAttributePointRequest(packet []byte) (SpendAttributePointRequestPacket, error) {
	var p SpendAttributePointRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// AttributePointSpentPacket carries an attribute's new value and the
// caster's remaining StatsPoints. It is sent by the server whenever a
// SpendAttributePointRequestPacket successfully resolves.
//
// MaxHealth/Health/MaxMana/Mana are included too, not just the touched
// attribute -- correction (août 2026): spending on Vitality/Energy also
// grows these (HeroStatsState.SpendAttributePoint), but the original
// packet only ever carried the attribute itself, so the client's own
// health/mana pools silently never grew even though the server's did.
type AttributePointSpentPacket struct {
	PlayerID    string `json:"playerId"`
	Attribute   int    `json:"attribute"`
	NewValue    int    `json:"newValue"`    // the attribute's new value after spending the point
	StatsPoints int    `json:"statsPoints"` // the caster's remaining attribute points
	MaxHealth   int    `json:"maxHealth"`
	Health      int    `json:"health"`
	MaxMana     int    `json:"maxMana"`
	Mana        int    `json:"mana"`
}

// CreateAttributePointSpentPacket returns a NetPacket which declares an
// AttributePointSpentPacket for the given player, attribute, its new
// value, the caster's remaining StatsPoints, and their current
// MaxHealth/Health/MaxMana/Mana.
func CreateAttributePointSpentPacket(
	playerID string, attribute, newValue, statsPoints, maxHealth, health, maxMana, mana int,
) (NetPacket, error) {
	spentPacket := AttributePointSpentPacket{
		PlayerID:    playerID,
		Attribute:   attribute,
		NewValue:    newValue,
		StatsPoints: statsPoints,
		MaxHealth:   maxHealth,
		Health:      health,
		MaxMana:     maxMana,
		Mana:        mana,
	}

	b, err := json.Marshal(spentPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.AttributePointSpent}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.AttributePointSpent,
		PacketData: b,
	}, nil
}

// UnmarshalAttributePointSpent unmarshals the given data to an
// AttributePointSpentPacket struct.
func UnmarshalAttributePointSpent(packet []byte) (AttributePointSpentPacket, error) {
	var p AttributePointSpentPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// RespecSingleAttributePointRequestPacket is sent by a client, requesting
// that one point previously spent on the given attribute be refunded
// (d2hero.HeroState.RespecSingleAttributePoint -- the attribute half of
// "Glyphe d'oubli").
type RespecSingleAttributePointRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	Attribute      int    `json:"attribute"`
}

// CreateRespecSingleAttributePointRequestPacket returns a NetPacket which
// declares a RespecSingleAttributePointRequestPacket for the given entity
// and attribute.
func CreateRespecSingleAttributePointRequestPacket(entityID string, attribute int) (NetPacket, error) {
	requestPacket := RespecSingleAttributePointRequestPacket{
		SourceEntityID: entityID,
		Attribute:      attribute,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.RespecSingleAttributePointRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.RespecSingleAttributePointRequest,
		PacketData: b,
	}, nil
}

// UnmarshalRespecSingleAttributePointRequest unmarshals the given data to a
// RespecSingleAttributePointRequestPacket struct.
func UnmarshalRespecSingleAttributePointRequest(packet []byte) (RespecSingleAttributePointRequestPacket, error) {
	var p RespecSingleAttributePointRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// SingleAttributePointRespecedPacket carries the attribute's new value
// after one point was refunded from it, and the caster's new StatsPoints
// total. It is sent by the server whenever a
// RespecSingleAttributePointRequestPacket successfully resolves.
//
// MaxHealth/Health/MaxMana/Mana included for the same reason as
// AttributePointSpentPacket's own -- refunding Vitality/Energy shrinks
// these too (HeroStatsState.RefundAttributePoint).
type SingleAttributePointRespecedPacket struct {
	PlayerID    string `json:"playerId"`
	Attribute   int    `json:"attribute"`
	NewValue    int    `json:"newValue"`
	StatsPoints int    `json:"statsPoints"`
	MaxHealth   int    `json:"maxHealth"`
	Health      int    `json:"health"`
	MaxMana     int    `json:"maxMana"`
	Mana        int    `json:"mana"`
}

// CreateSingleAttributePointRespecedPacket returns a NetPacket which
// declares a SingleAttributePointRespecedPacket for the given player,
// attribute, its new value, the caster's new StatsPoints total, and their
// current MaxHealth/Health/MaxMana/Mana.
func CreateSingleAttributePointRespecedPacket(
	playerID string, attribute, newValue, statsPoints, maxHealth, health, maxMana, mana int,
) (NetPacket, error) {
	respecedPacket := SingleAttributePointRespecedPacket{
		PlayerID:    playerID,
		Attribute:   attribute,
		NewValue:    newValue,
		StatsPoints: statsPoints,
		MaxHealth:   maxHealth,
		Health:      health,
		MaxMana:     maxMana,
		Mana:        mana,
	}

	b, err := json.Marshal(respecedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.SingleAttributePointRespeced}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.SingleAttributePointRespeced,
		PacketData: b,
	}, nil
}

// UnmarshalSingleAttributePointRespeced unmarshals the given data to a
// SingleAttributePointRespecedPacket struct.
func UnmarshalSingleAttributePointRespeced(packet []byte) (SingleAttributePointRespecedPacket, error) {
	var p SingleAttributePointRespecedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
