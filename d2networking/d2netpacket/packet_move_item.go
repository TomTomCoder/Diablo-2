package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// MoveToStashRequestPacket is sent by a client, requesting that the item
// at the given inventory slot be moved to the stash
// (d2hero.HeroState.MoveToStash).
type MoveToStashRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	InventoryIndex int    `json:"inventoryIndex"`
}

// CreateMoveToStashRequestPacket returns a NetPacket which declares a
// MoveToStashRequestPacket for the given entity and inventory slot.
func CreateMoveToStashRequestPacket(entityID string, inventoryIndex int) (NetPacket, error) {
	requestPacket := MoveToStashRequestPacket{
		SourceEntityID: entityID,
		InventoryIndex: inventoryIndex,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.MoveToStashRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.MoveToStashRequest,
		PacketData: b,
	}, nil
}

// UnmarshalMoveToStashRequest unmarshals the given data to a
// MoveToStashRequestPacket struct.
func UnmarshalMoveToStashRequest(packet []byte) (MoveToStashRequestPacket, error) {
	var p MoveToStashRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// ItemMovedToStashPacket carries the result of a successfully-resolved
// MoveToStashRequestPacket: the inventory slot that's now empty, and the
// stash slot the item landed in.
type ItemMovedToStashPacket struct {
	PlayerID       string `json:"playerId"`
	InventoryIndex int    `json:"inventoryIndex"`
	StashIndex     int    `json:"stashIndex"`
	ItemCode       string `json:"itemCode"`
}

// CreateItemMovedToStashPacket returns a NetPacket which declares an
// ItemMovedToStashPacket for the given player, slots, and item.
func CreateItemMovedToStashPacket(playerID string, inventoryIndex, stashIndex int, itemCode string) (NetPacket, error) {
	movedPacket := ItemMovedToStashPacket{
		PlayerID:       playerID,
		InventoryIndex: inventoryIndex,
		StashIndex:     stashIndex,
		ItemCode:       itemCode,
	}

	b, err := json.Marshal(movedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ItemMovedToStash}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ItemMovedToStash,
		PacketData: b,
	}, nil
}

// UnmarshalItemMovedToStash unmarshals the given data to an
// ItemMovedToStashPacket struct.
func UnmarshalItemMovedToStash(packet []byte) (ItemMovedToStashPacket, error) {
	var p ItemMovedToStashPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// MoveToInventoryRequestPacket is MoveToStashRequestPacket's mirror: moves
// an item from the stash back to the inventory
// (d2hero.HeroState.MoveToInventory).
type MoveToInventoryRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	StashIndex     int    `json:"stashIndex"`
}

// CreateMoveToInventoryRequestPacket returns a NetPacket which declares a
// MoveToInventoryRequestPacket for the given entity and stash slot.
func CreateMoveToInventoryRequestPacket(entityID string, stashIndex int) (NetPacket, error) {
	requestPacket := MoveToInventoryRequestPacket{
		SourceEntityID: entityID,
		StashIndex:     stashIndex,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.MoveToInventoryRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.MoveToInventoryRequest,
		PacketData: b,
	}, nil
}

// UnmarshalMoveToInventoryRequest unmarshals the given data to a
// MoveToInventoryRequestPacket struct.
func UnmarshalMoveToInventoryRequest(packet []byte) (MoveToInventoryRequestPacket, error) {
	var p MoveToInventoryRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// ItemMovedFromStashPacket carries the result of a successfully-resolved
// MoveToInventoryRequestPacket: the stash slot that's now empty, and the
// inventory slot the item landed in.
type ItemMovedFromStashPacket struct {
	PlayerID       string `json:"playerId"`
	StashIndex     int    `json:"stashIndex"`
	InventoryIndex int    `json:"inventoryIndex"`
	ItemCode       string `json:"itemCode"`
}

// CreateItemMovedFromStashPacket returns a NetPacket which declares an
// ItemMovedFromStashPacket for the given player, slots, and item.
func CreateItemMovedFromStashPacket(playerID string, stashIndex, inventoryIndex int, itemCode string) (NetPacket, error) {
	movedPacket := ItemMovedFromStashPacket{
		PlayerID:       playerID,
		StashIndex:     stashIndex,
		InventoryIndex: inventoryIndex,
		ItemCode:       itemCode,
	}

	b, err := json.Marshal(movedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ItemMovedFromStash}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ItemMovedFromStash,
		PacketData: b,
	}, nil
}

// UnmarshalItemMovedFromStash unmarshals the given data to an
// ItemMovedFromStashPacket struct.
func UnmarshalItemMovedFromStash(packet []byte) (ItemMovedFromStashPacket, error) {
	var p ItemMovedFromStashPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
