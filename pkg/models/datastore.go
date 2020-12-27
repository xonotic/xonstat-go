package models

import (
	"database/sql"
	"time"
)

// Datastore is the interface representing all database operations.
// This is useful for mocking or implementing different backends.
type Datastore interface {
	Begin() (*sql.Tx, error)

	// Server-oriented methods
	CServer(tx *sql.Tx, server Server) (int64, error)
	RServerByID(ID int) (*Server, error)
	RServersByHashkey(hashkey string) ([]*Server, error)
	RServersByName(name string) ([]*Server, error)
	UServer(tx *sql.Tx, server Server) error

	// Map-oriented methods
	CMap(tx *sql.Tx, m Map) (int64, error)
	RMapsByName(name string) ([]*Map, error)
	RMapByID(mapID int) (*Map, error)

	// Game-oriented methods
	CGame(tx *sql.Tx, game Game) (int64, error)
	RGamesByMatchID(matchID string) ([]*Game, error)
	RGameByID(gameID int) (*Game, error)
	RTeamGameStatsByGameID(gameID int) ([]*TeamGameStat, error)

	// Player-oriented methods
	CPlayer(tx *sql.Tx, player Player) (int64, error)
	RPlayersByHashkeyMulti(hashkeys []string) (map[string]*Player, error)
	RPlayerByID(playerID int) (*Player, error)
	RGameTypeSummariesByID(playerID int) ([]*PlayerGameTypeSummary, error)
	RGameIDsByPlayerID(playerID, limit int, gameTypeCd string) ([]int, error)
	RPlayerOverallStats(playerID int) ([]*PlayerOverallStats, error)
	UPlayer(tx *sql.Tx, player Player) error

	// Hashkey-oriented methods
	CHashkey(tx *sql.Tx, hashkey PlayerHashkey) error

	// PlayerGameStat-oriented methods
	CPlayerGameStat(tx *sql.Tx, pgs PlayerGameStat) (int64, error)
	RPlayerGameStatsByGameID(gameID int) ([]*PlayerGameStat, error)

	// PlayerWeaponStat-oriented methods
	CPlayerWeaponStat(tx *sql.Tx, pws PlayerWeaponStat) (int64, error)
	RPlayerWeaponStatsByGameID(gameID int) ([]*PlayerWeaponStat, error)
	RPlayerWeaponStatsByGameList(playerID int, gameIDs []int) ([]*PlayerWeaponStat, error)

	// WeaponInfo-oriented methods
	RWeaponInfoByGameID(gameID int) ([]*WeaponInfo, error)

	// TeamGameStat-oriented methods
	CTeamGameStat(tx *sql.Tx, tgs TeamGameStat) (int64, error)

	// PlayerGameFragMatrix-oriented methods
	CPlayerGameFragMatrix(tx *sql.Tx, fm PlayerGameFragMatrix) error
	RPlayerGameFragMatrixByGameID(gameID int) ([]*PlayerGameFragMatrix, error)

	// Summary stats by scope
	RSummaryStats(scope string) ([]*SummaryStat, error)

	// Top Players by time played
	RActivePlayers(limit, start int) ([]*ActivePlayer, error)
	RActivePlayersByServer(serverID int, cutoff *time.Time, limit int) ([]*ActivePlayer, error)
	RActivePlayersByMap(mapID int, cutoff *time.Time, limit int) ([]*ActivePlayer, error)

	// Top Servers by aggregate play time
	RActiveServers(limit, start int) ([]*ActiveServer, error)
	RActiveServersByMap(mapID int, cutoff *time.Time, limit int) ([]*ActiveServer, error)

	// Top Maps by times played
	RActiveMaps(limit, start int) ([]*ActiveMap, error)
	RActiveMapsByServer(serverID int, cutoff *time.Time, limit int) ([]*ActiveMap, error)

	// Top scoring players by server or map
	RServerActivePlayerScores(serverID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error)
	RMapActivePlayerScores(mapID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error)

	// RecentGames by various means
	RRecentGames(serverID int, mapID int, playerID int, gameTypeCd string, cutoff *time.Time, startGameID int, endGameID int, limit int) ([]*RecentGame, error)

	// PlayerGameNonParticipant-oriented methods
	CPlayerGameNonParticipant(tx *sql.Tx, pgnp PlayerGameNonParticipant) (int64, error)
	RPlayerGameNonParticipantsByGameID(gameID int) ([]*PlayerGameNonParticipant, error)

	// PlayerElo-oriented methods
	RPlayerElosByHashkey(hashkey string) ([]*PlayerElo, error)
}
