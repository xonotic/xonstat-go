package models

import (
	"database/sql"
	"time"
)

// Game is a single Xonotic match played.
type Game struct {
	GameID        int
	GameTypeCd    string
	GameTypeDescr string
	ServerID      int
	MapID         int
	Duration      *time.Duration // PostgreSQL interval
	Winner        sql.NullInt64
	MatchID       sql.NullString
	Mod           sql.NullString
	Players       []int
	StartDt       time.Time
	CreateDt      time.Time
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
	Nick             sql.NullString
	StrippedNick     sql.NullString
	Team             sql.NullInt32
	Rank             sql.NullInt32
	AliveTime        *time.Duration // PostgreSQL interval
	Kills            sql.NullInt32
	Deaths           sql.NullInt32
	Suicides         sql.NullInt32
	Score            sql.NullInt32
	Time             *time.Duration // PostgreSQL interval
	Captures         sql.NullInt32
	Pickups          sql.NullInt32
	Drops            sql.NullInt32
	Returns          sql.NullInt32
	Collects         sql.NullInt32
	Destroys         sql.NullInt32
	Pushes           sql.NullInt32
	CarrierFrags     sql.NullInt32
	EloDelta         sql.NullFloat64
	Fastest          *time.Duration // PostgreSQL interval
	AvgLatency       sql.NullFloat64
	TeamRank         sql.NullInt32
	ScoreboardPos    sql.NullInt32
	Laps             sql.NullInt32
	Revivals         sql.NullInt32
	Lives            sql.NullInt32
	CreateDt         time.Time
}

// NewPlayerGameStat creates a default record based on the "zeroed" stats of a particular game type.
func NewPlayerGameStat(gameTypeCd string) *PlayerGameStat {
	var pgs PlayerGameStat

	pgs.Score = sql.NullInt32{Valid: true, Int32: 0}

	switch gameTypeCd {
	case "as":
		pgs.Team = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Collects = sql.NullInt32{Valid: true, Int32: 0}

	case "ca", "dm", "duel", "rune", "tdm":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}

	case "ctf":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Captures = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Pickups = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Drops = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Returns = sql.NullInt32{Valid: true, Int32: 0}
		pgs.CarrierFrags = sql.NullInt32{Valid: true, Int32: 0}

	case "cts":
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}

	case "dom":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Pickups = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Drops = sql.NullInt32{Valid: true, Int32: 0}

	case "ft":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Revivals = sql.NullInt32{Valid: true, Int32: 0}

	case "ka":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Pickups = sql.NullInt32{Valid: true, Int32: 0}
		pgs.CarrierFrags = sql.NullInt32{Valid: true, Int32: 0}

	case "kh":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Pickups = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Captures = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Drops = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Pushes = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Destroys = sql.NullInt32{Valid: true, Int32: 0}
		pgs.CarrierFrags = sql.NullInt32{Valid: true, Int32: 0}

	case "nb":
		pgs.Kills = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Deaths = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Suicides = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Captures = sql.NullInt32{Valid: true, Int32: 0}
		pgs.Drops = sql.NullInt32{Valid: true, Int32: 0}
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
	Nick         sql.NullString
	StrippedNick sql.NullString
	Location     sql.NullString
	EmailAddr    sql.NullString
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
	Score          sql.NullInt32
	Rounds         sql.NullInt32
	Caps           sql.NullInt32
	CreateDt       time.Time
}

// NewTeamGameStat creates a "zeroed" TeamGameStat record depending upon the game type.
func NewTeamGameStat(gameTypeCd string) *TeamGameStat {
	var tgs TeamGameStat

	// All TeamGameStat records have a score
	tgs.Score = sql.NullInt32{Valid: true, Int32: 0}

	switch gameTypeCd {
	case "ca", "ft", "ka":
		tgs.Rounds = sql.NullInt32{Valid: true, Int32: 0}

	case "ctf":
		tgs.Caps = sql.NullInt32{Valid: true, Int32: 0}
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

// PlayerGameFragMatrix is a matrix of "who fragged who" in a given match.
type PlayerGameFragMatrix struct {
	GameID           int
	PlayerGameStatID int
	PlayerID         int
	PlayerIndex      int
	Matrix           map[int]int
}

// SummaryStat is a piece of the summary stat line at the top of the leaderboard page. It
// has a scope of either "all" or "day". Several of these are ultimately used to construct
// a string like:
// "Tracking X players and Y games (N dm; N ctf; N duel; N cts; N tdm; N other) since $DATE."
type SummaryStat struct {
	PlayerCount int
	GameTypeCd  string
	GameCount   int
	RefreshDt   time.Time
}

// ActivePlayer is a leaderboard item for a player's playing time over a specified
// time window.
type ActivePlayer struct {
	SortOrder int
	PlayerID  int
	Nick      string
	AliveTime time.Duration
	CreateDt  time.Time
}

// ActiveServer is a leaderboard item for a server's aggregate playing time over a specified
// time window.
type ActiveServer struct {
	SortOrder  int
	ServerID   int
	ServerName string
	PlayTime   time.Duration
	CreateDt   time.Time
}

// ActiveMap is a leaderboard item for the number of times a map was played over a specified
// time window.
type ActiveMap struct {
	SortOrder int
	MapID     int
	MapName   string
	Games     int
	CreateDt  time.Time
}

// RecentGame is an aggregated response for summary information from a match
type RecentGame struct {
	GameID          int
	GameTypeCd      string
	GameTypeDescr   string
	ServerID        int
	ServerName      string
	MapID           int
	MapName         string
	WinningTeam     sql.NullInt32
	WinningPlayerID int
	WinningNick     string
	CreateDt        time.Time
}

// PlayerGameNonParticipant houses people who were around for a match but didn't complete it.
type PlayerGameNonParticipant struct {
	PlayerGameNonParticipantID int
	PlayerID                   int
	GameID                     int
	Status                     string
	Nick                       sql.NullString
	StrippedNick               sql.NullString
	AliveTime                  *time.Duration // PostgreSQL interval
	Score                      sql.NullInt32
	CreateDt                   time.Time
}

// WeaponInfo is the weapon details of a single player weapon within a given game, paired
// with a little info about the player themself.
type WeaponInfo struct {
	PlayerWeaponStatID int
	PlayerID           int
	Nick               string
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

// PlayerGameTypeSummary holds the number of games for a player for a given game type,
// with wins, losses, and a win percentage thereof.
type PlayerGameTypeSummary struct {
	GameTypeCd string
	Wins       int
	Losses     int
}

// PlayerOverallStats holds aggregate stats about a player.
type PlayerOverallStats struct {
	GameTypeCd    string
	GameTypeDescr string
	Kills         sql.NullInt32
	Deaths        sql.NullInt32
	Pickups       sql.NullInt32
	Captures      sql.NullInt32
	CarrierFrags  sql.NullInt32
	LastPlayed    sql.NullTime
	TimePlayed    time.Duration
}
