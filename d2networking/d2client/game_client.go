package d2client

import (
	"fmt"
	"os"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapgen"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2math/d2vector"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapengine"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2clientconnectiontype"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2localclient"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2client/d2remoteclient"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket"
	"github.com/OpenDiablo2/OpenDiablo2/d2networking/d2netpacket/d2netpackettype"
	"github.com/OpenDiablo2/OpenDiablo2/d2script"
)

const logPrefix = "Game Client"

const (
	numSubtilesPerTile = 5
)

// GameClient manages a connection to d2server.GameServer
// and keeps a synchronized copy of the map and entities.
type GameClient struct {
	clientConnection ServerConnection                            // Abstract local/remote connection
	connectionType   d2clientconnectiontype.ClientConnectionType // Type of connection (local or remote)
	asset            *d2asset.AssetManager
	scriptEngine     *d2script.ScriptEngine
	GameState        *d2hero.HeroState              // local player state
	MapEngine        *d2mapengine.MapEngine         // Map and entities
	mapGen           *d2mapgen.MapGenerator         // map generator
	PlayerID         string                         // ID of the local player
	Players          map[string]*d2mapentity.Player // IDs of the other players
	Seed             int64                          // Map seed
	RegenMap         bool                           // Regenerate tile cache on render (map has changed)

	*d2util.Logger
}

// Create constructs a new GameClient and returns a pointer to it.
func Create(connectionType d2clientconnectiontype.ClientConnectionType,
	asset *d2asset.AssetManager,
	l d2util.LogLevel,
	scriptEngine *d2script.ScriptEngine) (*GameClient, error) {
	result := &GameClient{
		asset:          asset,
		MapEngine:      d2mapengine.CreateMapEngine(l, asset),
		Players:        make(map[string]*d2mapentity.Player),
		connectionType: connectionType,
		scriptEngine:   scriptEngine,
	}

	result.Logger = d2util.NewLogger()
	result.Logger.SetPrefix(logPrefix)
	result.Logger.SetLevel(l)

	// for a remote client connection, set loading to true - wait until we process the GenerateMapPacket
	// before we start updating map entites
	result.MapEngine.IsLoading = connectionType == d2clientconnectiontype.LANClient

	mapGen, err := d2mapgen.NewMapGenerator(asset, l, result.MapEngine)
	if err != nil {
		return nil, err
	}

	result.mapGen = mapGen

	switch connectionType {
	case d2clientconnectiontype.LANClient:
		result.clientConnection, err = d2remoteclient.Create(l, asset)
	case d2clientconnectiontype.LANServer:
		result.clientConnection, err = d2localclient.Create(asset, l, true)
	case d2clientconnectiontype.Local:
		result.clientConnection, err = d2localclient.Create(asset, l, false)
	default:
		err = fmt.Errorf("unknown client connection type specified: %d", connectionType)
	}

	if err != nil {
		return nil, err
	}

	result.clientConnection.SetClientListener(result)

	return result, nil
}

// Open creates the server and connects to it if the client is local.
// If the client is remote it sends a PlayerConnectionRequestPacket to the
// server (see d2netpacket).
func (g *GameClient) Open(connectionString, saveFilePath string) error {
	return g.clientConnection.Open(connectionString, saveFilePath)
}

// Close destroys the server if the client is local. For remote clients
// it sends a DisconnectRequestPacket (see d2netpacket).
func (g *GameClient) Close() error {
	return g.clientConnection.Close()
}

// Destroy does the same thing as Close.
func (g *GameClient) Destroy() error {
	return g.Close()
}

// OnPacketReceived is called by the ClientConection and processes incoming
// packets.
//
// nolint:gocyclo,gocognit,funlen // switch statement on packet type makes
// sense, no need to change. gocognit/funlen specifically found by
// golangci-lint (août 2026) once CI's own broken lint step got fixed for
// the first time this session -- same justification as the existing gocyclo
// suppression, this switch grows by one case per Devil packet type, not by
// real complexity.
func (g *GameClient) OnPacketReceived(packet d2netpacket.NetPacket) error {
	switch packet.PacketType {
	case d2netpackettype.GenerateMap:
		if err := g.handleGenerateMapPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.UpdateServerInfo:
		if err := g.handleUpdateServerInfoPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.AddPlayer:
		if err := g.handleAddPlayerPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.MovePlayer:
		if err := g.handleMovePlayerPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.CastSkill:
		if err := g.handleCastSkillPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SpawnItem:
		if err := g.handleSpawnItemPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.NPCHit:
		if err := g.handleNPCHitPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.NPCMoved:
		if err := g.handleNPCMovedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.NPCStatusEffect:
		if err := g.handleNPCStatusEffectPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.PlayerStatusEffect:
		if err := g.handlePlayerStatusEffectPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.PlayerDamaged:
		if err := g.handlePlayerDamagedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.GoldAwarded:
		if err := g.handleGoldAwardedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.ExperienceAwarded:
		if err := g.handleExperienceAwardedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.PlayerTeleported:
		if err := g.handlePlayerTeleportedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.PotionUsed:
		if err := g.handlePotionUsedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.ItemCrafted:
		if err := g.handleItemCraftedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SkillLearned:
		if err := g.handleSkillLearnedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SkillEquipped:
		if err := g.handleSkillEquippedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SkillsRespeced:
		if err := g.handleSkillsRespecedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SingleSkillRespeced:
		if err := g.handleSingleSkillRespecedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SkillPointInvested:
		if err := g.handleSkillPointInvestedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.AttributePointSpent:
		if err := g.handleAttributePointSpentPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.SingleAttributePointRespeced:
		if err := g.handleSingleAttributePointRespecedPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.Ping:
		if err := g.handlePingPacket(); err != nil {
			g.Errorf("GameClient: error responding to server ping: %s", err)
		}
	case d2netpackettype.PlayerDisconnectionNotification:
		if err := g.handlePlayerDisconnectionPacket(packet); err != nil {
			return err
		}
	case d2netpackettype.ServerClosed:
		// https://github.com/OpenDiablo2/OpenDiablo2/issues/802
		g.Infof("Server has been closed")
		os.Exit(0)
	case d2netpackettype.ServerFull:
		g.Infof("Server is full") // need to be verified
		os.Exit(0)
	default:
		g.Fatalf("Invalid packet type: %d", packet.PacketType)
	}

	return nil
}

// SendPacketToServer calls server.OnPacketReceived if the client is local.
// If it is remote the NetPacket sent over a UDP connection to the server.
func (g *GameClient) SendPacketToServer(packet d2netpacket.NetPacket) error {
	return g.clientConnection.SendPacketToServer(packet)
}

func (g *GameClient) handleGenerateMapPacket(packet d2netpacket.NetPacket) error {
	mapData, err := d2netpacket.UnmarshalGenerateMap(packet.PacketData)
	if err != nil {
		return err
	}

	if mapData.RegionType == d2enum.RegionAct1Town {
		g.mapGen.GenerateAct1Overworld()
	}

	g.RegenMap = true

	return nil
}

func (g *GameClient) handleUpdateServerInfoPacket(packet d2netpacket.NetPacket) error {
	serverInfo, err := d2netpacket.UnmarshalUpdateServerInfo(packet.PacketData)
	if err != nil {
		return err
	}

	g.MapEngine.SetSeed(serverInfo.Seed)
	g.PlayerID = serverInfo.PlayerID
	g.Seed = serverInfo.Seed
	g.Infof("Player id set to %s", serverInfo.PlayerID)

	return nil
}

func (g *GameClient) handleAddPlayerPacket(packet d2netpacket.NetPacket) error {
	player, err := d2netpacket.UnmarshalAddPlayer(packet.PacketData)
	if err != nil {
		return err
	}

	d2hero.HydrateSkills(player.Skills, g.asset)

	newPlayer := g.MapEngine.NewPlayer(player.ID, player.Name, player.X, player.Y, 0,
		player.HeroType, player.Stats, player.Skills, &player.Equipment, player.LeftSkill, player.RightSkill, player.Gold)

	g.Players[newPlayer.ID()] = newPlayer
	g.MapEngine.AddEntity(newPlayer)

	return nil
}

func (g *GameClient) handleSpawnItemPacket(packet d2netpacket.NetPacket) error {
	item, err := d2netpacket.UnmarshalSpawnItem(packet.PacketData)
	if err != nil {
		return err
	}

	itemEntity, err := g.MapEngine.NewItem(item.X, item.Y, item.Codes...)

	if err == nil {
		g.MapEngine.AddEntity(itemEntity)
	}

	return err
}

// Correction (août 2026): used to look up g.Players without a `found`
// check, same class of bug as handlePlayerDisconnectionPacket -- a
// MovePlayer packet for a player ID not (yet, or no longer) in g.Players
// would panic dereferencing a nil *d2mapentity.Player at player.SetPath.
func (g *GameClient) handleMovePlayerPacket(packet d2netpacket.NetPacket) error {
	movePlayer, err := d2netpacket.UnmarshalMovePlayer(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[movePlayer.PlayerID]
	if !found {
		return nil
	}

	start := d2vector.NewPositionTile(movePlayer.StartX, movePlayer.StartY)
	dest := d2vector.NewPositionTile(movePlayer.DestX, movePlayer.DestY)
	path := g.MapEngine.PathFind(start, dest)

	if len(path) > 0 {
		player.SetPath(path, func() {
			tilePosition := player.Position.Tile()
			tile := g.MapEngine.TileAt(int(tilePosition.X()), int(tilePosition.Y()))

			if tile == nil {
				return
			}

			player.SetIsInTown(tile.RegionType == d2enum.RegionAct1Town)

			err := player.SetAnimationMode(player.GetAnimationMode())

			if err != nil {
				fmtStr := "GameClient: error setting animation mode for player %s: %s"
				g.Errorf(fmtStr, player.ID(), err)
			}
		})
	}

	return nil
}

// handleNPCHitPacket applies a server-resolved hit to the local copy of the
// NPC: updates its HP, and removes it from the map if it died.
func (g *GameClient) handleNPCHitPacket(packet d2netpacket.NetPacket) error {
	hit, err := d2netpacket.UnmarshalNPCHit(packet.PacketData)
	if err != nil {
		return err
	}

	entity, found := g.MapEngine.Entities()[hit.EntityID]
	if !found {
		return nil
	}

	npc, ok := entity.(*d2mapentity.NPC)
	if !ok {
		return nil
	}

	npc.HP = hit.HP

	if hit.Died {
		g.MapEngine.RemoveEntity(npc)
	}

	return nil
}

// handleNPCMovedPacket applies a server-authoritative instant displacement
// (Télékinésie's Knockback, Vortex's Pull) to the local copy of the NPC.
func (g *GameClient) handleNPCMovedPacket(packet d2netpacket.NetPacket) error {
	moved, err := d2netpacket.UnmarshalNPCMoved(packet.PacketData)
	if err != nil {
		return err
	}

	entity, found := g.MapEngine.Entities()[moved.EntityID]
	if !found {
		return nil
	}

	npc, ok := entity.(*d2mapentity.NPC)
	if !ok {
		return nil
	}

	npc.Position.Set(moved.X, moved.Y)

	return nil
}

// handleNPCStatusEffectPacket applies a server-resolved timed status effect
// to the local copy of the NPC, via the same NPC methods the server itself
// used to apply it. Unrecognized effect names are silently ignored (forward
// compatibility with a server that knows about effects this client
// doesn't).
func (g *GameClient) handleNPCStatusEffectPacket(packet d2netpacket.NetPacket) error {
	status, err := d2netpacket.UnmarshalNPCStatusEffect(packet.PacketData)
	if err != nil {
		return err
	}

	entity, found := g.MapEngine.Entities()[status.EntityID]
	if !found {
		return nil
	}

	npc, ok := entity.(*d2mapentity.NPC)
	if !ok {
		return nil
	}

	switch status.Effect {
	case d2netpacket.NPCStatusSlowed:
		npc.ApplySlow(status.Until)
	case d2netpacket.NPCStatusImmobilized:
		npc.ApplyImmobilize(status.Until)
	case d2netpacket.NPCStatusAmplified:
		npc.ApplyAmplification(status.Until)
	case d2netpacket.NPCStatusResistanceStripped:
		npc.ApplyResistanceStrip(status.Until)
	}

	return nil
}

// handlePlayerStatusEffectPacket applies a server-resolved change to the
// given player's own status effect (Bouclier de mana/Armure de glace
// toggles, Éveil du Nexus's timed immunity) to the local copy of their
// stats. Unrecognized effect names are silently ignored.
func (g *GameClient) handlePlayerStatusEffectPacket(packet d2netpacket.NetPacket) error {
	status, err := d2netpacket.UnmarshalPlayerStatusEffect(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[status.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	switch status.Effect {
	case d2netpacket.PlayerStatusManaShield:
		player.Stats.ManaShieldActive = status.Active
	case d2netpacket.PlayerStatusArmureDeGlace:
		player.Stats.ArmureDeGlaceActive = status.Active
	case d2netpacket.PlayerStatusMagicImmune:
		player.Stats.ApplyMagicImmunity(status.Until)
	}

	return nil
}

// handlePlayerDamagedPacket applies a server-resolved monster hit to the
// local copy of the given player's stats (their HP bar, wherever the HUD
// reads it, follows automatically since it's the same *HeroStatsState).
//
// ponytail: no death handling (no game-over screen, no respawn) -- Health
// just floors at 0. See ROADMAP.md Phase 4.
func (g *GameClient) handlePlayerDamagedPacket(packet d2netpacket.NetPacket) error {
	hit, err := d2netpacket.UnmarshalPlayerDamaged(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[hit.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	player.Stats.Health = hit.HP
	player.Stats.Mana = hit.Mana

	return nil
}

// handleGoldAwardedPacket applies a server-resolved gold total to the
// local copy of the given player's entity.
func (g *GameClient) handleGoldAwardedPacket(packet d2netpacket.NetPacket) error {
	awarded, err := d2netpacket.UnmarshalGoldAwarded(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[awarded.PlayerID]
	if !found {
		return nil
	}

	player.Gold = awarded.Gold

	return nil
}

// handlePotionUsedPacket applies a server-resolved Mana total to the local
// copy of the given player's stats, after they consumed a potion.
func (g *GameClient) handlePotionUsedPacket(packet d2netpacket.NetPacket) error {
	used, err := d2netpacket.UnmarshalPotionUsed(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[used.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	player.Stats.Mana = used.Mana

	return nil
}

// handleItemCraftedPacket applies a server-resolved Cube de Nexus craft to
// the local copy of the given player: the new RightHand item (Craft only
// ever touches that slot) and their remaining Gold.
func (g *GameClient) handleItemCraftedPacket(packet d2netpacket.NetPacket) error {
	crafted, err := d2netpacket.UnmarshalItemCrafted(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[crafted.PlayerID]
	if !found {
		return nil
	}

	if player.Equipment == nil || player.Equipment.RightHand == nil {
		return nil
	}

	player.Equipment.RightHand.ItemCode = crafted.OutputItemCode
	player.Equipment.RightHand.ItemName = crafted.OutputItemName
	player.Gold = crafted.Gold

	return nil
}

// handleSkillLearnedPacket adds a server-resolved learned skill to the
// local copy of the given player's Skills, and updates their remaining
// SkillPoints.
func (g *GameClient) handleSkillLearnedPacket(packet d2netpacket.NetPacket) error {
	learned, err := d2netpacket.UnmarshalSkillLearned(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[learned.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	skill := d2hero.NewDevilHeroSkill(learned.SkillID)
	if skill == nil {
		return nil
	}

	if player.Skills == nil {
		player.Skills = make(map[int]*d2hero.HeroSkill)
	}

	player.Skills[learned.SkillID] = skill
	player.Stats.SkillPoints = learned.SkillPoints

	return nil
}

// handleSkillEquippedPacket assigns the local copy of the given player's
// LeftSkill/RightSkill to their already-known skill of the given ID. Does
// nothing if the skill isn't known locally (the server would have refused
// the request in that case anyway).
func (g *GameClient) handleSkillEquippedPacket(packet d2netpacket.NetPacket) error {
	equipped, err := d2netpacket.UnmarshalSkillEquipped(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[equipped.PlayerID]
	if !found {
		return nil
	}

	skill, known := player.Skills[equipped.SkillID]
	if !known {
		return nil
	}

	switch d2hero.SkillSlot(equipped.Slot) {
	case d2hero.SkillSlotLeft:
		player.LeftSkill = skill
	case d2hero.SkillSlotRight:
		player.RightSkill = skill
	}

	return nil
}

// handleSkillsRespecedPacket clears the local copy of the given player's
// Skills (and equipped Left/RightSkill, since they'd otherwise reference a
// skill that no longer exists), and updates SkillPoints.
func (g *GameClient) handleSkillsRespecedPacket(packet d2netpacket.NetPacket) error {
	respeced, err := d2netpacket.UnmarshalSkillsRespeced(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[respeced.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	player.Skills = make(map[int]*d2hero.HeroSkill)
	player.LeftSkill = nil
	player.RightSkill = nil
	player.Stats.SkillPoints = respeced.SkillPoints

	// Correction (août 2026): RespecSkills also refunds every attribute
	// point ever spent (HeroStatsState.RespecAllAttributePoints) -- the
	// packet used to carry only SkillPoints, so the client never learned
	// about this other half of the same "Respec partiel" action.
	player.Stats.StatsPoints = respeced.StatsPoints
	player.Stats.Strength = respeced.Strength
	player.Stats.Energy = respeced.Energy
	player.Stats.Dexterity = respeced.Dexterity
	player.Stats.Vitality = respeced.Vitality
	player.Stats.MaxHealth = respeced.MaxHealth
	player.Stats.Health = respeced.Health
	player.Stats.MaxMana = respeced.MaxMana
	player.Stats.Mana = respeced.Mana

	return nil
}

// handleSingleSkillRespecedPacket removes one specific skill from the
// local copy of the given player's Skills, clearing Left/RightSkill only
// if either was equipped to that skill, and updates SkillPoints.
func (g *GameClient) handleSingleSkillRespecedPacket(packet d2netpacket.NetPacket) error {
	respeced, err := d2netpacket.UnmarshalSingleSkillRespeced(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[respeced.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	delete(player.Skills, respeced.SkillID)

	if player.LeftSkill != nil && player.LeftSkill.SkillRecord != nil && player.LeftSkill.SkillRecord.ID == respeced.SkillID {
		player.LeftSkill = nil
	}

	if player.RightSkill != nil && player.RightSkill.SkillRecord != nil && player.RightSkill.SkillRecord.ID == respeced.SkillID {
		player.RightSkill = nil
	}

	player.Stats.SkillPoints = respeced.SkillPoints

	return nil
}

// handleSkillPointInvestedPacket updates the local copy of the given
// player's already-known skill with its new invested-points total, and
// updates their remaining SkillPoints. A no-op if the skill isn't known
// locally yet -- the server is the source of truth and would have rejected
// the request in that case too (HeroState.InvestSkillPoint).
func (g *GameClient) handleSkillPointInvestedPacket(packet d2netpacket.NetPacket) error {
	invested, err := d2netpacket.UnmarshalSkillPointInvested(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[invested.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	skill, known := player.Skills[invested.SkillID]
	if !known {
		return nil
	}

	skill.SkillPoints = invested.InvestedPoints
	player.Stats.SkillPoints = invested.SkillPoints

	return nil
}

// setAttributeValue writes newValue to stats' Strength/Energy/Dexterity/
// Vitality field matching attr. A no-op for an unrecognized attribute --
// mirrors GameServer.attributeValue's own switch on the server side.
func setAttributeValue(stats *d2hero.HeroStatsState, attr, newValue int) {
	switch d2hero.Attribute(attr) {
	case d2hero.AttributeStrength:
		stats.Strength = newValue
	case d2hero.AttributeEnergy:
		stats.Energy = newValue
	case d2hero.AttributeDexterity:
		stats.Dexterity = newValue
	case d2hero.AttributeVitality:
		stats.Vitality = newValue
	}
}

// handleAttributePointSpentPacket applies a server-resolved attribute-point
// spend to the local copy of the given player's stats.
func (g *GameClient) handleAttributePointSpentPacket(packet d2netpacket.NetPacket) error {
	spent, err := d2netpacket.UnmarshalAttributePointSpent(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[spent.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	setAttributeValue(player.Stats, spent.Attribute, spent.NewValue)
	player.Stats.StatsPoints = spent.StatsPoints

	// Correction (août 2026): spending on Vitality/Energy also grows
	// MaxHealth/MaxMana (HeroStatsState.SpendAttributePoint) -- the packet
	// used to carry only the touched attribute, so the client's own health/
	// mana pools never grew even though the server's did.
	player.Stats.MaxHealth = spent.MaxHealth
	player.Stats.Health = spent.Health
	player.Stats.MaxMana = spent.MaxMana
	player.Stats.Mana = spent.Mana

	return nil
}

// handleSingleAttributePointRespecedPacket applies a server-resolved
// single-attribute-point refund to the local copy of the given player's
// stats.
func (g *GameClient) handleSingleAttributePointRespecedPacket(packet d2netpacket.NetPacket) error {
	respeced, err := d2netpacket.UnmarshalSingleAttributePointRespeced(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[respeced.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	setAttributeValue(player.Stats, respeced.Attribute, respeced.NewValue)
	player.Stats.StatsPoints = respeced.StatsPoints

	// Correction (août 2026): same reasoning as
	// handleAttributePointSpentPacket -- refunding Vitality/Energy shrinks
	// MaxHealth/MaxMana too.
	player.Stats.MaxHealth = respeced.MaxHealth
	player.Stats.Health = respeced.Health
	player.Stats.MaxMana = respeced.MaxMana
	player.Stats.Mana = respeced.Mana

	return nil
}

// handleExperienceAwardedPacket applies a server-resolved experience/level
// state to the local copy of the given player's stats.
func (g *GameClient) handleExperienceAwardedPacket(packet d2netpacket.NetPacket) error {
	awarded, err := d2netpacket.UnmarshalExperienceAwarded(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[awarded.PlayerID]
	if !found || player.Stats == nil {
		return nil
	}

	player.Stats.Experience = awarded.Experience
	player.Stats.Level = awarded.Level
	player.Stats.SkillPoints = awarded.SkillPoints
	player.Stats.StatsPoints = awarded.StatsPoints

	return nil
}

// handlePlayerTeleportedPacket snaps the local copy of the given player
// directly to its server-resolved position (Téléportation) -- unlike
// handleMovePlayerPacket, this doesn't path a walk between two points.
func (g *GameClient) handlePlayerTeleportedPacket(packet d2netpacket.NetPacket) error {
	teleported, err := d2netpacket.UnmarshalPlayerTeleported(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[teleported.PlayerID]
	if !found {
		return nil
	}

	player.Position.Set(teleported.X, teleported.Y)

	return nil
}

// Correction (août 2026): used to look up g.Players without a `found`
// check, same class of bug as handlePlayerDisconnectionPacket -- a
// CastSkill packet for a player ID not in g.Players would panic
// dereferencing a nil *d2mapentity.Player at player.StopMoving.
func (g *GameClient) handleCastSkillPacket(packet d2netpacket.NetPacket) error {
	playerCast, err := d2netpacket.UnmarshalCast(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[playerCast.SourceEntityID]
	if !found {
		return nil
	}

	player.StopMoving()

	castX := playerCast.TargetX * numSubtilesPerTile
	castY := playerCast.TargetY * numSubtilesPerTile

	direction := player.Position.DirectionTo(*d2vector.NewVector(castX, castY))
	player.SetDirection(direction)

	skillRecord := g.asset.Records.Skill.Details[playerCast.SkillID]

	// Correction (août 2026): Skill.Details is the stock skills.txt table
	// this client-side visual-effect code was originally written against
	// -- none of Devil's own skills (SkillTraitDeFeu etc.) are in it, so
	// skillRecord is nil for every one of them. Every line below this
	// dereferences it (skillRecord.Cltmissile inside createMissileEntities,
	// skillRecord.Summon, skillRecord.Anim), a guaranteed nil-pointer panic
	// the instant a real player actually cast a Devil skill -- never caught
	// before because no UI trigger had ever sent a real CastPacket with a
	// Devil skill ID (the one existing test using a Devil skill ID targets
	// an *unknown* player, returning above before reaching this code).
	// Server-side combat resolution (resolveMeleeHit) already works
	// correctly for Devil skills entirely independent of this table --
	// this only gates client-side missile/summon/cast-animation visuals,
	// which have nothing to render yet regardless (no Devil skill sprites
	// exist, ROADMAP.md Phase 6), so skipping them silently costs nothing
	// real today.
	if skillRecord == nil {
		return nil
	}

	missileEntities, err := g.createMissileEntities(skillRecord, player, castX, castY)
	if err != nil {
		return err
	}

	var summonedNpcEntity *d2mapentity.NPC
	if skillRecord.Summon != "" {
		summonedNpcEntity, err = g.createSummonedNpcEntity(skillRecord, int(castX), int(castY))

		if err != nil {
			return err
		}
	}

	player.StartCasting(skillRecord.Anim, func() {
		if len(missileEntities) > 0 {
			// shoot the missiles of the skill after the player has finished casting
			for _, missileEntity := range missileEntities {
				g.MapEngine.AddEntity(missileEntity)
			}
		}

		if summonedNpcEntity != nil {
			// summon the referenced NPC after the player has finished casting
			g.MapEngine.AddEntity(summonedNpcEntity)
		}
	})

	overlayRecord := g.asset.Records.Layout.Overlays[skillRecord.Castoverlay]

	return g.playCastOverlay(overlayRecord, int(player.Position.X()), int(player.Position.Y()))
}

func (g *GameClient) createSummonedNpcEntity(skillRecord *d2records.SkillRecord, x, y int) (*d2mapentity.NPC, error) {
	monsterStatsRecord := g.asset.Records.Monster.Stats[skillRecord.Summon]

	if monsterStatsRecord == nil {
		fmtErr := "cannot cast skill - No monstat entry for \"%s\""
		return nil, fmt.Errorf(fmtErr, skillRecord.Summon)
	}

	// https://github.com/OpenDiablo2/OpenDiablo2/issues/803
	summonedNpcEntity, err := g.MapEngine.NewNPC(x, y, monsterStatsRecord, 0)
	if err != nil {
		return nil, err
	}

	return summonedNpcEntity, nil
}

func (g *GameClient) createMissileEntities(
	skillRecord *d2records.SkillRecord,
	player *d2mapentity.Player,
	castX, castY float64,
) ([]*d2mapentity.Missile, error) {
	missileRecords := []*d2records.MissileRecord{
		g.asset.Records.GetMissileByName(skillRecord.Cltmissile),
		g.asset.Records.GetMissileByName(skillRecord.Cltmissilea),
		g.asset.Records.GetMissileByName(skillRecord.Cltmissileb),
		g.asset.Records.GetMissileByName(skillRecord.Cltmissilec),
		g.asset.Records.GetMissileByName(skillRecord.Cltmissiled),
	}

	missileEntities := make([]*d2mapentity.Missile, 0)

	for _, missileRecord := range missileRecords {
		if missileRecord == nil {
			continue
		}

		missileEntity, err := g.createMissileEntity(missileRecord, player, castX, castY)
		if err != nil {
			return nil, err
		}

		missileEntities = append(missileEntities, missileEntity)
	}

	return missileEntities, nil
}

func (g *GameClient) createMissileEntity(
	missileRecord *d2records.MissileRecord,
	player *d2mapentity.Player,
	castX, castY float64,
) (*d2mapentity.Missile, error) {
	if missileRecord == nil {
		return nil, nil
	}

	radians := d2math.GetRadiansBetween(
		player.Position.X(),
		player.Position.Y(),
		castX,
		castY,
	)

	missileEntity, err := g.MapEngine.NewMissile(
		int(player.Position.X()),
		int(player.Position.Y()),
		g.asset.Records.Missiles[missileRecord.Id],
	)

	if err != nil {
		return nil, err
	}

	missileEntity.SetRadians(radians, func() {
		g.MapEngine.RemoveEntity(missileEntity)
	})

	return missileEntity, nil
}

func (g *GameClient) playCastOverlay(overlayRecord *d2records.OverlayRecord, x, y int) error {
	if overlayRecord == nil {
		return nil
	}

	overlayEntity, err := g.MapEngine.NewCastOverlay(
		x,
		y,
		overlayRecord,
	)
	if err != nil {
		return err
	}

	overlayEntity.SetOnDoneFunc(func() {
		g.MapEngine.RemoveEntity(overlayEntity)
	})

	g.MapEngine.AddEntity(overlayEntity)

	return nil
}

func (g *GameClient) handlePingPacket() error {
	pongPacket, err := d2netpacket.CreatePongPacket(g.PlayerID)
	if err != nil {
		return err
	}

	err = g.clientConnection.SendPacketToServer(pongPacket)

	if err != nil {
		return err
	}

	return nil
}

// Correction (août 2026): used to look up g.Players without a `found`
// check -- for a duplicate/replayed disconnection notification (this
// packet type is broadcast unconditionally by GameServer with no
// de-dup, and the game supports a UDP client connection, which can
// legitimately duplicate datagrams), the second delivery finds nothing
// left in g.Players, so `player` is a nil *d2mapentity.Player. Passing
// that nil pointer into MapEngine.RemoveEntity(entity d2interface.MapEntity)
// wraps it in a non-nil interface value (the interface carries the
// *Player type even though the pointer itself is nil) -- RemoveEntity's
// own `entity == nil` guard doesn't catch this classic Go gotcha, so it
// falls through to entity.ID(), a nil-pointer dereference.
func (g *GameClient) handlePlayerDisconnectionPacket(packet d2netpacket.NetPacket) error {
	disconnectPacket, err := d2netpacket.UnmarshalPlayerDisconnectionRequest(packet.PacketData)
	if err != nil {
		return err
	}

	player, found := g.Players[disconnectPacket.ID]
	if !found {
		return nil
	}

	g.MapEngine.RemoveEntity(player)
	delete(g.Players, disconnectPacket.ID)

	return nil
}

// IsSinglePlayer returns a bool for whether the game is a single-player game
func (g *GameClient) IsSinglePlayer() bool {
	return g.connectionType == d2clientconnectiontype.Local
}
