package d2netpackettype

import (
	"encoding/json"
)

// NetPacketType is an enum referring to all packet types in package
// d2netpacket.
type NetPacketType uint32

// (Except NetPacket which declares a NetPacketType to specify the packet body
// type. See d2netpackettype.NetPacket.)
//
// # Warning
//
// Do NOT re-arrange the order of these packet values unless you want to
// break compatibility between clients of slightly different versions.
// Also note that the packet id is a byte, so if we use more than 256 of
// these then we are doing something very wrong.
const (
	UpdateServerInfo                  NetPacketType = iota // Sent by the server, client sets the given player ID and map seed
	GenerateMap                                            // Sent by the server, client generates a map
	AddPlayer                                              // Sent by the server, client adds a player
	MovePlayer                                             // Sent by client or server, moves a player entity
	PlayerConnectionRequest                                // Sent by the remote client when connecting
	PlayerDisconnectionNotification                        // Sent by the remote client when disconnecting
	Ping                                                   // Requests a Pong packet
	Pong                                                   // Responds to a Ping packet
	ServerClosed                                           // Sent by the local host when it has closed the server
	CastSkill                                              // Sent by client or server, indicates entity casting skill
	SpawnItem                                              // Sent by server
	SavePlayer                                             // Sent by the client, saves the player
	ServerFull                                             // Sent by server when server has reached max connections
	NPCHit                                                 // Sent by the server, an NPC took damage (and possibly died)
	PlayerDamaged                                          // Sent by the server, a player took damage (and possibly died)
	GoldAwarded                                            // Sent by the server, a player's gold total changed
	ExperienceAwarded                                      // Sent by the server, a player gained experience (and possibly leveled up)
	PlayerTeleported                                       // Sent by the server, a player was instantly moved (Devil's own "Téléportation" skill)
	UsePotionRequest                                       // Sent by the client, requests using the potion in a given belt slot
	PotionUsed                                             // Sent by the server, a player's Mana changed from using a potion
	LearnSkillRequest                                      // Sent by the client, requests learning a given skill
	SkillLearned                                           // Sent by the server, a player learned a skill (and spent a skill point)
	RespecSkillsRequest                                    // Sent by the client, requests forgetting every learned skill (Respec partiel)
	SkillsRespeced                                         // Sent by the server, a player's skills were all forgotten and refunded
	RespecSingleSkillRequest                               // Sent by the client, requests forgetting one specific skill (Glyphe d'oubli)
	SingleSkillRespeced                                    // Sent by the server, one of a player's skills was forgotten and refunded
	InvestSkillPointRequest                                // Sent by the client, requests investing another point into an already-known skill
	SkillPointInvested                                     // Sent by the server, a player invested another point into a known skill
	SpendAttributePointRequest                             // Sent by the client, requests spending a level-up point on a given attribute
	AttributePointSpent                                    // Sent by the server, a player spent a point on an attribute
	RespecSingleAttributePointRequest                      // Sent by the client, requests refunding one point from a given attribute (Glyphe d'oubli)
	SingleAttributePointRespeced                           // Sent by the server, one attribute point was refunded
	NPCMoved                                               // Sent by the server, an NPC was instantly displaced (Télékinésie/Vortex)
	NPCStatusEffect                                        // Sent by the server, a timed status effect was applied to an NPC (slow/immobilize/amplify/resistance strip)

	UnknownPacketType = 666
)

func (n NetPacketType) String() string {
	strings := map[NetPacketType]string{
		UpdateServerInfo:                  "UpdateServerInfo",
		GenerateMap:                       "GenerateMap",
		AddPlayer:                         "AddPlayer",
		MovePlayer:                        "MovePlayer",
		PlayerConnectionRequest:           "PlayerConnectionRequest",
		PlayerDisconnectionNotification:   "PlayerDisconnectionNotification",
		Ping:                              "Ping",
		Pong:                              "Pong",
		ServerClosed:                      "ServerClosed",
		CastSkill:                         "CastSkill",
		SpawnItem:                         "SpawnItem",
		SavePlayer:                        "SavePlayer",
		ServerFull:                        "ServerFull",
		NPCHit:                            "NPCHit",
		PlayerDamaged:                     "PlayerDamaged",
		GoldAwarded:                       "GoldAwarded",
		ExperienceAwarded:                 "ExperienceAwarded",
		PlayerTeleported:                  "PlayerTeleported",
		UsePotionRequest:                  "UsePotionRequest",
		PotionUsed:                        "PotionUsed",
		LearnSkillRequest:                 "LearnSkillRequest",
		SkillLearned:                      "SkillLearned",
		RespecSkillsRequest:               "RespecSkillsRequest",
		SkillsRespeced:                    "SkillsRespeced",
		RespecSingleSkillRequest:          "RespecSingleSkillRequest",
		SingleSkillRespeced:               "SingleSkillRespeced",
		InvestSkillPointRequest:           "InvestSkillPointRequest",
		SkillPointInvested:                "SkillPointInvested",
		SpendAttributePointRequest:        "SpendAttributePointRequest",
		AttributePointSpent:               "AttributePointSpent",
		RespecSingleAttributePointRequest: "RespecSingleAttributePointRequest",
		SingleAttributePointRespeced:      "SingleAttributePointRespeced",
		NPCMoved:                          "NPCMoved",
		NPCStatusEffect:                   "NPCStatusEffect",
	}

	return strings[n]
}

// MarshalPacket marshals the packet to a byte slice
func (n NetPacketType) MarshalPacket() ([]byte, error) {
	p, err := json.Marshal(n)
	if err != nil {
		return p, err
	}

	return p, nil
}
