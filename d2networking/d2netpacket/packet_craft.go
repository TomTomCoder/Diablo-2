package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// CraftRequestPacket is sent by a client, requesting that a Cube de Nexus
// recipe be crafted (d2hero.HeroState.Craft).
type CraftRequestPacket struct {
	SourceEntityID string `json:"sourceEntityId"`
	RecipeID       string `json:"recipeId"`
}

// CreateCraftRequestPacket returns a NetPacket which declares a
// CraftRequestPacket for the given entity and recipe.
func CreateCraftRequestPacket(entityID, recipeID string) (NetPacket, error) {
	requestPacket := CraftRequestPacket{
		SourceEntityID: entityID,
		RecipeID:       recipeID,
	}

	b, err := json.Marshal(requestPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.CraftRequest}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.CraftRequest,
		PacketData: b,
	}, nil
}

// UnmarshalCraftRequest unmarshals the given data to a CraftRequestPacket
// struct.
func UnmarshalCraftRequest(packet []byte) (CraftRequestPacket, error) {
	var p CraftRequestPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}

// ItemCraftedPacket carries the recipe's output item (now equipped in the
// caster's RightHand -- HeroState.Craft only ever touches that slot, see
// its own doc comment) and the caster's remaining Gold. Sent by the server
// whenever a CraftRequestPacket successfully resolves.
type ItemCraftedPacket struct {
	PlayerID       string `json:"playerId"`
	RecipeID       string `json:"recipeId"`
	OutputItemCode string `json:"outputItemCode"`
	OutputItemName string `json:"outputItemName"`
	Gold           int    `json:"gold"`
}

// CreateItemCraftedPacket returns a NetPacket which declares an
// ItemCraftedPacket for the given player, recipe, output item, and
// remaining Gold.
func CreateItemCraftedPacket(playerID, recipeID, outputItemCode, outputItemName string, gold int) (NetPacket, error) {
	craftedPacket := ItemCraftedPacket{
		PlayerID:       playerID,
		RecipeID:       recipeID,
		OutputItemCode: outputItemCode,
		OutputItemName: outputItemName,
		Gold:           gold,
	}

	b, err := json.Marshal(craftedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ItemCrafted}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ItemCrafted,
		PacketData: b,
	}, nil
}

// UnmarshalItemCrafted unmarshals the given data to an ItemCraftedPacket
// struct.
func UnmarshalItemCrafted(packet []byte) (ItemCraftedPacket, error) {
	var p ItemCraftedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
