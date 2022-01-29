package submission

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"gitlab.com/xonotic/xonstat/pkg/models"
)

var ErrDuplicateGame = fmt.Errorf("duplicate game")

// ShouldUpdateServer determines if the database server record needs to be updated with
// new information coming from the submission.
func ShouldUpdateServer(incoming, existing *models.Server) bool {
	if incoming.Name.Valid && incoming.Name.String != existing.Name.String {
		return true
	}

	if incoming.HashKey.Valid && incoming.HashKey.String != existing.HashKey.String {
		return true
	}

	if incoming.IPAddr.Valid && incoming.IPAddr.String != existing.IPAddr.String {
		return true
	}

	if incoming.Port.Valid && incoming.Port.Int64 != existing.Port.Int64 {
		return true
	}

	if incoming.Revision.Valid && incoming.Revision.String != existing.Revision.String {
		return true
	}

	if incoming.ImpureCvars.Valid && incoming.ImpureCvars.Int64 != existing.ImpureCvars.Int64 {
		return true
	}

	return false
}

// GetOrCreateServer finds an existing server matching the one provided or constructs a new one.
func GetOrCreateServer(tx *sql.Tx, db models.Datastore, rawServer *models.Server) (*models.Server, error) {
	var servers []*models.Server
	var err error

	if rawServer.HashKey.Valid {
		log.Printf("Looking for server by hashkey '%s'", rawServer.HashKey.String)
		servers, err = db.RServersByHashkey(rawServer.HashKey.String)
		if err != nil {
			return nil, err
		}
	}

	// Fall back to searching by name if hashkey is not provided.
	if len(servers) == 0 && rawServer.Name.Valid {
		log.Printf("Looking for server by name '%s'", rawServer.Name.String)
		servers, err = db.RServersByName(rawServer.Name.String)
		if err != nil {
			return nil, err
		}
	}

	if len(servers) == 0 {
		// We haven't found a matching server. Create one.
		serverID, err := db.CServer(tx, *rawServer)
		if err != nil {
			return nil, err
		}
		rawServer.ServerID = int(serverID)
		log.Printf("Created new server %d.", serverID)
		return rawServer, nil
	}

	if len(servers) == 1 {
		log.Printf("Found matching server %d.", servers[0].ServerID)
	} else {
		log.Printf("Multiple matching servers found. Using the first one (%d).", servers[0].ServerID)
	}
	rawServer.ServerID = servers[0].ServerID

	if ShouldUpdateServer(rawServer, servers[0]) {
		log.Printf("Updating server %d.", rawServer.ServerID)
		err := db.UServer(tx, *rawServer)
		if err != nil {
			return nil, err
		}
	}

	// Do not proceed if a server is inactive (banned, broken, etc).
	if !servers[0].ActiveInd {
		return nil, fmt.Errorf("server %d is inactive", servers[0].ServerID)
	}

	return servers[0], nil
}

// GetOrCreateMap finds an existing map matching the one provided or constructs a new one.
func GetOrCreateMap(tx *sql.Tx, db models.Datastore, rawMap *models.Map) (*models.Map, error) {
	var maps []*models.Map
	var err error

	log.Printf("Looking for map by name '%s'", rawMap.Name)
	maps, err = db.RMapsByName(rawMap.Name)
	if err != nil {
		return nil, err
	}

	if len(maps) == 0 {
		// We haven't found a matching map. Create one.
		mapID, err := db.CMap(tx, *rawMap)
		if err != nil {
			return nil, err
		}
		rawMap.MapID = int(mapID)
		log.Printf("Created new map %d.", mapID)
		return rawMap, nil
	}

	if len(maps) == 1 {
		log.Printf("Found matching map %d.", maps[0].MapID)
	} else {
		log.Printf("Multiple matching maps found. Using the first one (%d).", maps[0].MapID)
	}
	rawMap.MapID = maps[0].MapID

	return maps[0], nil
}

// CreateGame creates a game record in the database, first checking if it exists using the MatchID.
// We expect a game to be inserted upon each submission, so this method only returns an error.
func CreateGame(tx *sql.Tx, db models.Datastore, s *Submission) error {
	if s.Game.MatchID.Valid {
		games, err := db.RGamesByMatchID(s.Game.ServerID, s.Game.MatchID.String)
		if err != nil {
			return err
		}

		if len(games) > 0 {
			log.Printf("A game with match_id %s for server %d already exists in the database.",
				s.Game.MatchID.String, s.Game.ServerID)
			return ErrDuplicateGame
		}
	}

	// Identify non-participants by their pids.
	nonPartipants := make(map[int]struct{})
	for _, np := range s.NonParticipants {
		nonPartipants[np.PlayerID] = struct{}{}
	}

	// For easier queries later, we store the PIDs right on the game entry as an array.
	var humansInGame []int
	for _, player := range s.Players {
		_, isNonParticipant := nonPartipants[player.PlayerID]
		if player.PlayerID > 2 && !isNonParticipant {
			humansInGame = append(humansInGame, player.PlayerID)
		}
	}
	s.Game.Players = humansInGame

	gameID, err := db.CGame(tx, *s.Game)
	if err != nil {
		return err
	}
	s.Game.GameID = int(gameID)
	log.Printf("Created game %d from match %s.", gameID, s.Game.MatchID.String)

	// Update the next record along the way.
	for _, pgs := range s.PlayerGameStats {
		pgs.GameID = int(gameID)
	}

	return nil
}

// ShouldUpdatePlayer determines if the incoming data has a new piece of information
// that should be persisted to the database with an update.
func ShouldUpdatePlayer(incoming, existing *models.Player) bool {
	if incoming.Nick.Valid && incoming.Nick.String != existing.Nick.String {
		return true
	}
	// TODO: register a nick change, if that is something we still want...
	return false
}

// GetOrCreatePlayers fetches existing players or creates new ones based upon the data
// in the submission. This one is done in batch to reduce SQL calls.
func GetOrCreatePlayers(tx *sql.Tx, db models.Datastore, s *Submission) (map[string]*models.Player, error) {
	// These records are fixed for bots and anons (untracked players) respectively
	bot := models.Player{PlayerID: 1, Nick: sql.NullString{Valid: true, String: "bot"}}
	anon := models.Player{PlayerID: 2, Nick: sql.NullString{Valid: true, String: "Anonymous Player"}}

	// This is the final return value. All player records found or created in the database.
	playersByHashkey := make(map[string]*models.Player)

	var hashkeys []string                      // used for searching
	hashkeySet := make(map[string]struct{}, 0) // used for keeping track of who we haven't processed

	// Bots and untracked players need no fetches from the database.
	for hashkey := range s.PlayersByHashkey {
		if strings.HasPrefix(hashkey, "bot#") {
			// bot
			playersByHashkey[hashkey] = &bot
		} else if strings.HasPrefix(hashkey, "player#") {
			// untracked player
			playersByHashkey[hashkey] = &anon
		} else {
			// human that we need to look for/create
			hashkeys = append(hashkeys, hashkey)
			hashkeySet[hashkey] = struct{}{}
		}
	}

	// Players that already exist are more complicated. We first fetch who we can,
	// then update them if need be.
	if len(hashkeys) > 0 {
		playersByHashkeyDB, err := db.RPlayersByHashkeyMulti(hashkeys)
		if err != nil {
			return nil, err
		}

		for hashkey, dbPlayer := range playersByHashkeyDB {
			rawPlayer := s.PlayersByHashkey[hashkey]
			if ShouldUpdatePlayer(rawPlayer, dbPlayer) {
				// Apply the update to the DB value...
				dbPlayer.Nick = rawPlayer.Nick
				dbPlayer.StrippedNick = rawPlayer.StrippedNick

				// ...and save it to the database.
				db.UPlayer(tx, *dbPlayer)
				log.Printf("Updated player %d '%s'", dbPlayer.PlayerID, dbPlayer.StrippedNick.String)
			}

			playersByHashkey[hashkey] = dbPlayer

			delete(hashkeySet, hashkey) // done processing this one
		}
	}

	// The remaining players left in hashkeySet need to be created.
	for hashkey := range hashkeySet {
		newPlayer := s.PlayersByHashkey[hashkey]

		playerID, err := db.CPlayer(tx, *newPlayer)
		if err != nil {
			return nil, err
		}

		err = db.CHashkey(tx, models.PlayerHashkey{Hashkey: hashkey, PlayerID: int(playerID)})
		if err != nil {
			return nil, err
		}

		newPlayer.PlayerID = int(playerID)
		playersByHashkey[hashkey] = newPlayer

		log.Printf("Created player %d '%s'", playerID, newPlayer.StrippedNick.String)
	}

	// Reflect the new PIDs in the submission accordingly (at least for the next table being modified).
	for hashkey, player := range playersByHashkey {
		*s.PlayersByHashkey[hashkey] = *player

		pgs, ok := s.PlayerGameStatsByHashkey[hashkey]
		if ok {
			pgs.PlayerID = player.PlayerID
		}

		pgnp, ok := s.NonParticipantsByHashkey[hashkey]
		if ok {
			pgnp.PlayerID = player.PlayerID
		}
	}

	return playersByHashkey, nil
}

// CreatePlayerGameStats inserts all of the game stat records to the database.
func CreatePlayerGameStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for _, pgs := range s.PlayerGameStats {
		pgsID, err := db.CPlayerGameStat(tx, *pgs)
		if err != nil {
			return err
		}
		pgs.PlayerGameStatID = int(pgsID)
	}

	return nil
}

// CreateNonParticipants inserts all of the non-participant records to the database.
func CreateNonParticipants(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for _, pgnp := range s.NonParticipants {
		pgnp.GameID = s.Game.GameID

		id, err := db.CPlayerGameNonParticipant(tx, *pgnp)
		if err != nil {
			return err
		}
		pgnp.PlayerGameNonParticipantID = int(id)
	}

	return nil
}

// CreatePlayerWeaponStats inserts all of the weapon stat records to the database.
func CreatePlayerWeaponStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for hashkey, pwsList := range s.PlayerWeaponStatsByHashkey {
		for _, pws := range pwsList {
			pws.PlayerID = s.PlayersByHashkey[hashkey].PlayerID

			// We don't store weapon information for bots.
			if pws.PlayerID == 1 {
				break
			}

			pws.GameID = s.Game.GameID
			pws.PlayerGameStatID = s.PlayerGameStatsByHashkey[hashkey].PlayerGameStatID

			pwsID, err := db.CPlayerWeaponStat(tx, *pws)
			if err != nil {
				return err
			}

			pws.PlayerWeaponStatID = int(pwsID)
		}
	}

	return nil
}

// CreateTeamGameStats inserts all of the team game stat records to the database.
func CreateTeamGameStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for _, tgs := range s.TeamGameStats {
		tgs.GameID = s.Game.GameID
		tgsID, err := db.CTeamGameStat(tx, *tgs)
		if err != nil {
			return err
		}

		tgs.TeamGameStatID = int(tgsID)
	}

	return nil
}

// ShouldDoFragMatrix determines if we're going to process frag matrix stuff or not.
func ShouldDoFragMatrix(gameTypeCd string) bool {
	switch gameTypeCd {
	case "as", "ca", "ctf", "dm", "dom", "ft", "freezetag", "ka", "kh", "rune", "tdm":
		return true
	}

	return false
}

// CreateFragMatrix inserts all of the frag matrix records to the database.
func CreateFragMatrix(tx *sql.Tx, db models.Datastore, s *Submission) error {
	switch s.Game.GameTypeCd {
	case "as", "ca", "ctf", "dm", "dom", "ft", "freezetag", "ka", "kh", "rune", "tdm":
	default:
		return nil
	}

	for playerIndex, matrix := range s.FragMatrixByIndex {
		hashkey := s.HashkeysByIndex[playerIndex]
		pgsID := s.PlayerGameStatsByHashkey[hashkey].PlayerGameStatID
		pid := s.PlayersByHashkey[hashkey].PlayerID

		fm := models.PlayerGameFragMatrix{
			GameID:           s.Game.GameID,
			PlayerGameStatID: pgsID,
			PlayerID:         pid,
			PlayerIndex:      playerIndex,
			Matrix:           matrix,
		}

		err := db.CPlayerGameFragMatrix(tx, fm)
		if err != nil {
			return err
		}
	}

	return nil
}

// Submit takes a fully-formed submission and stores it in the database, filling out all the
// missing information (like primary key values) along the way.
func Submit(s *Submission, db models.Datastore) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	server, err := GetOrCreateServer(tx, db, s.Server)
	if err != nil {
		tx.Rollback()
		return err
	}
	s.Game.ServerID = server.ServerID

	m, err := GetOrCreateMap(tx, db, s.Map)
	if err != nil {
		tx.Rollback()
		return err
	}
	s.Game.MapID = m.MapID

	_, err = GetOrCreatePlayers(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = CreateGame(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = CreatePlayerGameStats(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = CreateNonParticipants(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = CreatePlayerWeaponStats(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = CreateTeamGameStats(tx, db, s)
	if err != nil {
		tx.Rollback()
		return err
	}

	if ShouldDoFragMatrix(s.Game.GameTypeCd) {
		err = CreateFragMatrix(tx, db, s)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
