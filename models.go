package models

import (
	"time"
)

type Game struct {
	GameId     int
	GameTypeCd string
	ServerId   int
	MapId      int
	Duration   time.Duration
	Winner     int8
	MatchId    string
	Mod        string
	Category   string
	Players    []int
	StartDt    time.Time
	CreateDt   time.Time
}

type Server struct {
	ServerId    int
	Name        string
	Location    string
	IPAddr      string
	Port        int
	HashKey     string
	PublicKey   string
	Revision    string
	PureInd     bool
	ImpureCvars int
	EloInd      bool
	Categories  []string
	ActiveInd   bool
	CreateDt    time.Time
}

type Map struct {
	MapId    int
	Name     string
	Version  int
	Pk3Name  string
	CurlUrl  string
	CreateDt time.Time
}

type PlayerGameStat struct {
	PlayerGameStatId int
	PlayerId         int
	GameId           int
	Nick             string
	StrippedNick     string
	Team             int
	Rank             int
	AliveTime        time.Duration
	Kills            int
	Deaths           int
	Suicides         int
	Score            int
	Time             time.Duration
	Captures         int
	Pickups          int
	Drops            int
	Returns          int
	Collects         int
	Destroys         int
	Pushes           int
	CarrierFrags     int
	EloDelta         float64
	Fastest          time.Duration
	AvgLatency       float64
	TeamRank         int
	ScoreboardPos    int
	Laps             int
	Revivals         int
	Lives            int
	CreateDt         time.Time
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
