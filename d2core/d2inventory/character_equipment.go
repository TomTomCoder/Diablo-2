package d2inventory

// CharacterEquipment stores equipments of a character
type CharacterEquipment struct {
	Head      *InventoryItemArmor  `json:"head"`      // Head
	Torso     *InventoryItemArmor  `json:"torso"`     // TR
	Legs      *InventoryItemArmor  `json:"legs"`      // Legs
	RightArm  *InventoryItemArmor  `json:"rightArm"`  // RA
	LeftArm   *InventoryItemArmor  `json:"leftArm"`   // LA
	LeftHand  *InventoryItemWeapon `json:"leftHand"`  // LH
	RightHand *InventoryItemWeapon `json:"rightHand"` // RH
	Shield    *InventoryItemArmor  `json:"shield"`    // SH
	// S1-S8?

	// Amulet has no D2 equivalent field here (the original OpenDiablo2
	// equipment set never modeled a neck slot) -- added for Devil, whose
	// starting equipment includes an amulet (devil_mage_character_design.md
	// §5: "Pendentif Arcane"). InventoryItemMisc is reused rather than a
	// new item type, since it already has everything a non-weapon,
	// non-armor equippable needs (ItemCode + nil-safe GetItemCode).
	Amulet *InventoryItemMisc `json:"amulet"`
}
