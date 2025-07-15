package player

import (
	"sort"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/util"
)

// GameTypeSummaryBase is for keeping track of the number of games played by a player plus win:loss ratio.
type GameTypeSummaryBase struct {
	GameTypeCd string
	Games      int
	Wins       int
	Losses     int
	WinRatio   float32
}

// ByGames implements sort.Interface for []GameTypeSummaryBase based on the Games field.
type ByGames []*GameTypeSummaryBase

func (a ByGames) Len() int           { return len(a) }
func (a ByGames) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGames) Less(i, j int) bool { return a[i].Games > a[j].Games } // descending order

// GameTypeSummaryData retrieves the summary data for a player by ID.
// Among the list of returned summaries is also an "overall" entry that is the sum value for all types.
func GameTypeSummaryData(db models.Datastore, playerID int) ([]*GameTypeSummaryBase, error) {
	rawSummaries, err := db.RGameTypeSummariesByID(playerID)
	if err != nil {
		return nil, err
	}

	// This is a "meta" type that sums all of the others
	overall := GameTypeSummaryBase{GameTypeCd: "overall"}

	var summaries []*GameTypeSummaryBase
	for _, rs := range rawSummaries {
		s := GameTypeSummaryBase{
			GameTypeCd: rs.GameTypeCd,
			Games:      rs.Wins + rs.Losses,
			Wins:       rs.Wins,
			Losses:     rs.Losses,
		}

		// Exclude FFA gametypes (CTS, DM, KA, MAYHEM, LMS)
		switch rs.GameTypeCd {
		case "dm", "cts", "ka", "keepaway", "mayhem", "lms":
			s.Wins = 0
			s.Losses = 0
			s.WinRatio = 0.0
		default:
			s.WinRatio = util.Percentage(rs.Wins, rs.Wins+rs.Losses)
		}

		// A running tally of the counts we've seen so far
		overall.Games += s.Games
		overall.Wins += s.Wins
		overall.Losses += s.Losses

		summaries = append(summaries, &s)
	}

	overall.WinRatio = util.Percentage(overall.Wins, overall.Wins+overall.Losses)

	// Put the "overall" entry first in the list so it sorts in a stable fashion.
	summaries = append([]*GameTypeSummaryBase{&overall}, summaries...)

	sort.Sort(ByGames(summaries))

	return summaries, nil
}

// OverallStatsBase is for aggregate stats (kills, deaths, etc)
type OverallStatsBase struct {
	GameTypeCd    string
	GameTypeDescr string
	Kills         int
	Deaths        int
	KDRatio       float32
	Captures      int
	Pickups       int
	CapRatio      float32
	CarrierFrags  int
	LastPlayed    models.MultiDt
	TimePlayed    models.MultiDuration
}

// OverallStatsData retrieves aggregate stats and converts them to a view-agnostic form.
func OverallStatsData(db models.Datastore, playerID int) ([]*OverallStatsBase, error) {
	rawOverallStats, err := db.RPlayerOverallStats(playerID)
	if err != nil {
		return nil, err
	}

	overallStats := make([]*OverallStatsBase, 0, len(rawOverallStats))
	for _, elem := range rawOverallStats {
		lastPlayed, err := models.NewMultiDt(elem.LastPlayed.Time)
		if err != nil {
			return nil, err
		}

		timePlayed := models.NewMultiDuration(elem.TimePlayed)

		stat := OverallStatsBase{
			GameTypeCd:    elem.GameTypeCd,
			GameTypeDescr: elem.GameTypeDescr,
			Kills:         int(elem.Kills.Int32),
			Deaths:        int(elem.Deaths.Int32),
			KDRatio:       util.Ratio(int(elem.Kills.Int32), int(elem.Deaths.Int32)),
			Captures:      int(elem.Captures.Int32),
			Pickups:       int(elem.Pickups.Int32),
			CapRatio:      util.Ratio(int(elem.Captures.Int32), int(elem.Pickups.Int32)),
			CarrierFrags:  int(elem.CarrierFrags.Int32),
			LastPlayed:    *lastPlayed,
			TimePlayed:    *timePlayed,
		}

		overallStats = append(overallStats, &stat)
	}

	return overallStats, nil
}

// InfoBase is the view-agnostic representation of player information.
type InfoBase struct {
	PlayerID  int
	Nick      *models.MultiNick
	ActiveInd bool
	CakeDayInd bool
	CreateDt  *models.MultiDt
}

func InfoData(db models.Datastore, playerID int) (*InfoBase, error) {
	rawPlayer, err := db.RPlayerByID(playerID)
	if err != nil {
		return nil, err
	}

	nick := models.NewMultiNick(rawPlayer.Nick.String)
	dt, err := models.NewMultiDt(rawPlayer.CreateDt)
	if err != nil {
		return nil, err
	}

	cakeDay := false
	if util.IsAnniversary(dt.Dt, time.Now()) {
		cakeDay = true
	}

	return &InfoBase{
		PlayerID:  rawPlayer.PlayerID,
		Nick:      nick,
		ActiveInd: rawPlayer.ActiveInd,
		CakeDayInd: cakeDay,
		CreateDt:  dt,
	}, nil
}
