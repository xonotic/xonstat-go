package models

import (
	"database/sql"
)

// Datastore is the interface representing all database operations.
// This is useful for mocking or implementing different backends.
type Datastore interface {
	Begin() (*sql.Tx, error)

	// Server-oriented methods
	CServer(tx *sql.Tx, server Server) (int64, error)
	RServersByHashkey(hashkey string) ([]*Server, error)
	RServersByName(name string) ([]*Server, error)
	UServer(tx *sql.Tx, server Server) error

	// Map-oriented methods
	CMap(tx *sql.Tx, m Map) (int64, error)
	RMapsByName(name string) ([]*Map, error)

	// Game-oriented methods
	CGame(tx *sql.Tx, game Game) (int64, error)
	RGamesByMatchID(matchID string) ([]*Game, error)

	// Player-oriented methods
	CPlayer(tx *sql.Tx, player Player) (int64, error)
	RPlayersByHashkeyMulti(hashkeys []string) (map[string]*Player, error)
	UPlayer(tx *sql.Tx, player Player) error

	// Hashkey-oriented methods
	CHashkey(tx *sql.Tx, hashkey PlayerHashkey) error

	// PlayerGameStat-oriented methods
	CPlayerGameStat(tx *sql.Tx, pgs PlayerGameStat) (int64, error)

	// PlayerWeaponStat-oriented methods
	CPlayerWeaponStat(tx *sql.Tx, pws PlayerWeaponStat) (int64, error)

	// TeamGameStat-oriented methods
	CTeamGameStat(tx *sql.Tx, tgs TeamGameStat) (int64, error)

	// PlayerGameFragMatrix-oriented methods
	CPlayerGameFragMatrix(tx *sql.Tx, fm PlayerGameFragMatrix) error
}
