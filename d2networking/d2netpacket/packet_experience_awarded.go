package d2netpacket

import (
	"encoding/json"

	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
)

// ExperienceAwardedPacket carries a player's new experience/level state. It
// is sent by the server whenever experience is granted (e.g. a monster's
// death).
type ExperienceAwardedPacket struct {
	PlayerID    string `json:"playerId"`
	Experience  int    `json:"experience"`
	Level       int    `json:"level"`
	SkillPoints int    `json:"skillPoints"`
	StatsPoints int    `json:"statsPoints"`
}

// CreateExperienceAwardedPacket returns a NetPacket which declares an
// ExperienceAwardedPacket with the given player's new experience state.
func CreateExperienceAwardedPacket(playerID string, experience, level, skillPoints, statsPoints int) (NetPacket, error) {
	experienceAwardedPacket := ExperienceAwardedPacket{
		PlayerID:    playerID,
		Experience:  experience,
		Level:       level,
		SkillPoints: skillPoints,
		StatsPoints: statsPoints,
	}

	b, err := json.Marshal(experienceAwardedPacket)
	if err != nil {
		return NetPacket{PacketType: d2netpackettype.ExperienceAwarded}, err
	}

	return NetPacket{
		PacketType: d2netpackettype.ExperienceAwarded,
		PacketData: b,
	}, nil
}

// UnmarshalExperienceAwarded unmarshals the given data to an
// ExperienceAwardedPacket struct
func UnmarshalExperienceAwarded(packet []byte) (ExperienceAwardedPacket, error) {
	var p ExperienceAwardedPacket
	if err := json.Unmarshal(packet, &p); err != nil {
		return p, err
	}

	return p, nil
}
