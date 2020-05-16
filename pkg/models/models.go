package models

import (
	"database/sql"
	"time"
)

type Game struct {
	GameId     int
	GameTypeCd string
	ServerId   int
	MapId      int
	Duration   *time.Duration
	Winner     *int
	MatchId    *string
	Mod        *string
	Category   *string
	Players    []int
	StartDt    time.Time
	CreateDt   time.Time
}

type Server struct {
	ServerId    int
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

type Map struct {
	MapId    int
	Name     string
	Version  int
	Pk3Name  sql.NullString
	CurlUrl  sql.NullString
	CreateDt time.Time
}

type PlayerGameStat struct {
	PlayerGameStatId int
	PlayerId         int
	GameId           int
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

type PlayerWeaponStat struct {
	PlayerWeaponStatId int
	PlayerId           int
	GameId             int
	PlayerGameStatId   int
	WeaponCd           string
	Actual             int
	Max                int
	Hit                int
	Fired              int
	Frags              int
	CreateDt           time.Time
}

type Player struct {
	PlayerId     int
	Nick         *string
	StrippedNick *string
	Location     *string
	EmailAddr    *string
	ActiveInd    bool
	CreateDt     time.Time
}

type PlayerHashkey struct {
	PlayerId  int
	Hashkey   string
	ActiveInd bool
	CreateDt  time.Time
}

type TeamGameStat struct {
	TeamGameStatId int
	GameId         int
	Team           int
	Score          *int
	Rounds         *int
	Caps           *int
	CreateDt       time.Time
}

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

type PlayerGameAnticheat struct {
	PlayerGameAnticheatID int
	PlayerID              int
	GameID                int
	Key                   string
	Value                 float64
	CreateDt              time.Time
}
