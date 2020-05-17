package models

import (
	"database/sql"
	"time"
)

// Game is a single Xonotic match played.
type Game struct {
	GameID     int
	GameTypeCd string
	ServerID   int
	MapID      int
	Duration   *time.Duration
	Winner     *int
	MatchID    *string
	Mod        *string
	Category   *string
	Players    []int
	StartDt    time.Time
	CreateDt   time.Time
}

// Server is a Xonotic server which hosts games.
type Server struct {
	ServerID    int
	Name        sql.NullString
	Location    sql.NullString
	IPAddr      sql.NullString
	Port        sql.NullInt64
	HashKey     sql.NullString
	PublicKey   sql.NullString
	Revision    sql.NullString
	PureInd     bool
	ImpureCvars sql.NullInt64
	EloInd      bool
	Categories  []string
	ActiveInd   bool
	CreateDt    time.Time
}

// Map is the arena games are played in.
type Map struct {
	MapID    int
	Name     string
	Version  int
	Pk3Name  sql.NullString
	CurlURL  sql.NullString
	CreateDt time.Time
}

// PlayerGameStat houses the statistics for a single player (bot or human) in a game.
type PlayerGameStat struct {
	PlayerGameStatID int
	PlayerID         int
	GameID           int
	Nick             *string
	StrippedNick     *string
	Team             *int
	Rank             *int
	AliveTime        *time.Duration
	Kills            *int
	Deaths           *int
	Suicides         *int
	Score            *int
	Time             *time.Duration
	Captures         *int
	Pickups          *int
	Drops            *int
	Returns          *int
	Collects         *int
	Destroys         *int
	Pushes           *int
	CarrierFrags     *int
	EloDelta         *float64
	Fastest          *time.Duration
	AvgLatency       *float64
	TeamRank         *int
	ScoreboardPos    *int
	Laps             *int
	Revivals         *int
	Lives            *int
	CreateDt         time.Time
}

// NewPlayerGameStat creates a default record based on the "zeroed" stats of a particular game type.
func NewPlayerGameStat(gameTypeCd string) *PlayerGameStat {
	var pgs PlayerGameStat

	score := 0
	pgs.Score = &score

	switch gameTypeCd {
	case "as":
		team, kills, deaths, suicides, collects := 0, 0, 0, 0, 0
		pgs.Team, pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Collects = &team, &kills, &deaths,
			&suicides, &collects

	case "ca", "dm", "duel", "rune", "tdm":
		kills, deaths, suicides := 0, 0, 0
		pgs.Kills, pgs.Deaths, pgs.Suicides = &kills, &deaths, &suicides

	case "ctf":
		kills, captures, pickups, drops, returns, carrierFrags := 0, 0, 0, 0, 0, 0
		pgs.Kills, pgs.Captures, pgs.Pickups, pgs.Drops, pgs.Returns, pgs.CarrierFrags = &kills,
			&captures, &pickups, &drops, &returns, &carrierFrags

	case "cts":
		deaths := 0
		pgs.Deaths = &deaths

	case "dom":
		kills, deaths, suicides, pickups, drops := 0, 0, 0, 0, 0
		pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Pickups, pgs.Drops = &kills, &deaths, &suicides,
			&pickups, &drops

	case "ft":
		kills, deaths, suicides, revivals := 0, 0, 0, 0
		pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Revivals = &kills, &deaths, &suicides, &revivals

	case "ka":
		kills, deaths, suicides, pickups, carrierFrags := 0, 0, 0, 0, 0
		pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Pickups, pgs.CarrierFrags = &kills, &deaths,
			&suicides, &pickups, &carrierFrags

	case "kh":
		kills, deaths, suicides, pickups, captures := 0, 0, 0, 0, 0
		drops, pushes, destroys, carrierFrags := 0, 0, 0, 0

		pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Pickups, pgs.Captures = &kills, &deaths,
			&suicides, &pickups, &captures

		pgs.Drops, pgs.Pushes, pgs.Destroys, pgs.CarrierFrags = &drops, &pushes, &destroys,
			&carrierFrags

	case "nb":
		kills, deaths, suicides, captures, drops := 0, 0, 0, 0, 0
		pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Captures, pgs.Drops = &kills, &deaths,
			&suicides, &captures, &drops
	}

	return &pgs
}

// PlayerWeaponStat is the weapon details of a single player weapon within a given game.
type PlayerWeaponStat struct {
	PlayerWeaponStatID int
	PlayerID           int
	GameID             int
	PlayerGameStatID   int
	WeaponCd           string
	Actual             int
	Max                int
	Hit                int
	Fired              int
	Frags              int
	CreateDt           time.Time
}

// Player is a bot or human that participated in a game at some point.
type Player struct {
	PlayerID     int
	Nick         *string
	StrippedNick *string
	Location     *string
	EmailAddr    *string
	ActiveInd    bool
	CreateDt     time.Time
}

// PlayerHashkey is an identifier for a Player. A single player may have multiple hashkeys
// depending upon how many installations of Xonotic they have. A hashkey corresponds to
// the key_0.d0si file in the Xonotic user directory.
type PlayerHashkey struct {
	PlayerID  int
	Hashkey   string
	ActiveInd bool
	CreateDt  time.Time
}

// TeamGameStat holds team-specific stats for a game: number of team flag captures in CTF or
// rounds won in CA, for example.
type TeamGameStat struct {
	TeamGameStatID int
	GameID         int
	Team           int
	Score          *int
	Rounds         *int
	Caps           *int
	CreateDt       time.Time
}

// NewTeamGameStat creates a "zeroed" TeamGameStat record depending upon the game type.
func NewTeamGameStat(gameTypeCd string) *TeamGameStat {
	var tgs TeamGameStat

	// All TeamGameStat records have a score
	score := 0
	tgs.Score = &score

	switch gameTypeCd {
	case "ca", "ft", "ka":
		rounds := 0
		tgs.Rounds = &rounds

	case "ctf":
		caps := 0
		tgs.Caps = &caps
	}

	return &tgs
}

// PlayerGameAnticheat holds telemetry from the anticheat subsystem in the Xonotic server.
type PlayerGameAnticheat struct {
	PlayerGameAnticheatID int
	PlayerID              int
	GameID                int
	Key                   string
	Value                 float64
	CreateDt              time.Time
}
