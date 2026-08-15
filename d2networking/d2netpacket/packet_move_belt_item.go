package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// MoveToBeltRequestPacket is sent by a client, requesting that the potion
// at the given inventory slot be moved to the belt
// (d2hero.HeroState.MoveToBelt).
type MoveToBeltRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	InventoryIndex int    `json:"inventoryIndex"`
}

// CreateMoveToBeltRequestPacket returns a NetPacket which declares a
// MoveToBeltRequestPacket for the given entity and inventory slot.
func CreateMoveToBeltRequestPacket(entityID string, inventoryIndex int) (NetPacket, error) {
	requestPacket := MoveToBeltRequestPacket{
		SourceEntityID: entityID,
		InventoryIndex: inventoryIndex,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.MoveToBeltRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.MoveToBeltRequest,
		PacketData: b,
	}, nil
}

// UnmarshalMoveToBeltRequest unmarshals the given data to a
// MoveToBeltRequestPacket struct.
func UnmarshalMoveToBeltRequest(packet []byte) (MoveToBeltRequestPacket, error) {
	var p MoveToBeltRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// ItemMovedToBeltPacket carries the result of a successfully-resolved
// MoveToBeltRequestPacket: the inventory slot that's now empty, and the
// belt slot the potion landed in.
type ItemMovedToBeltPacket struct {
	PlayerID       string `json:"playerId"`
	InventoryIndex int    `json:"inventoryIndex"`
	BeltIndex      int    `json:"beltIndex"`
	ItemCode       string `json:"itemCode"`
}

// CreateItemMovedToBeltPacket returns a NetPacket which declares an
// ItemMovedToBeltPacket for the given player, slots, and item.
func CreateItemMovedToBeltPacket(playerID string, inventoryIndex, beltIndex int, itemCode string) (NetPacket, error) {
	movedPacket := ItemMovedToBeltPacket{
		PlayerID:       playerID,
		InventoryIndex: inventoryIndex,
		BeltIndex:      beltIndex,
		ItemCode:       itemCode,
	}

	b, err := json.Marshal(movedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ItemMovedToBelt}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ItemMovedToBelt,
		PacketData: b,
	}, nil
}

// UnmarshalItemMovedToBelt unmarshals the given data to an
// ItemMovedToBeltPacket struct.
func UnmarshalItemMovedToBelt(packet []byte) (ItemMovedToBeltPacket, error) {
	var p ItemMovedToBeltPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// MoveFromBeltRequestPacket is MoveToBeltRequestPacket's mirror: moves a
// potion from the belt back to the inventory
// (d2hero.HeroState.MoveFromBelt).
type MoveFromBeltRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	BeltIndex      int    `json:"beltIndex"`
}

// CreateMoveFromBeltRequestPacket returns a NetPacket which declares a
// MoveFromBeltRequestPacket for the given entity and belt slot.
func CreateMoveFromBeltRequestPacket(entityID string, beltIndex int) (NetPacket, error) {
	requestPacket := MoveFromBeltRequestPacket{
		SourceEntityID: entityID,
		BeltIndex:      beltIndex,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.MoveFromBeltRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.MoveFromBeltRequest,
		PacketData: b,
	}, nil
}

// UnmarshalMoveFromBeltRequest unmarshals the given data to a
// MoveFromBeltRequestPacket struct.
func UnmarshalMoveFromBeltRequest(packet []byte) (MoveFromBeltRequestPacket, error) {
	var p MoveFromBeltRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// ItemMovedFromBeltPacket carries the result of a successfully-resolved
// MoveFromBeltRequestPacket: the belt slot that's now empty, and the
// inventory slot the potion landed in.
type ItemMovedFromBeltPacket struct {
	PlayerID       string `json:"playerId"`
	BeltIndex      int    `json:"beltIndex"`
	InventoryIndex int    `json:"inventoryIndex"`
	ItemCode       string `json:"itemCode"`
}

// CreateItemMovedFromBeltPacket returns a NetPacket which declares an
// ItemMovedFromBeltPacket for the given player, slots, and item.
func CreateItemMovedFromBeltPacket(playerID string, beltIndex, inventoryIndex int, itemCode string) (NetPacket, error) {
	movedPacket := ItemMovedFromBeltPacket{
		PlayerID:       playerID,
		BeltIndex:      beltIndex,
		InventoryIndex: inventoryIndex,
		ItemCode:       itemCode,
	}

	b, err := json.Marshal(movedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ItemMovedFromBelt}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ItemMovedFromBelt,
		PacketData: b,
	}, nil
}

// UnmarshalItemMovedFromBelt unmarshals the given data to an
// ItemMovedFromBeltPacket struct.
func UnmarshalItemMovedFromBelt(packet []byte) (ItemMovedFromBeltPacket, error) {
	var p ItemMovedFromBeltPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
