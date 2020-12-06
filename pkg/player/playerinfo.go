package player

import (
	"sort"
	"gitlab.com/xonotic/xonstat/pkg/models"
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
func (a ByGames) Less(i, j int) bool { return a[i].Games > a[j].Games }  // descending order

func calcPct(numerator, denominator int) float32 {
	if denominator == 0 {
		denominator = 1
	}

	return float32(numerator) / float32(denominator) * 100.0
}

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
			Games:  rs.Wins + rs.Losses,
			Wins:   rs.Wins,
			Losses: rs.Losses,
		}

		switch rs.GameTypeCd {
		case "dm", "cts", "ka", "keepaway":
			s.WinRatio = 0.0
		default:
			s.WinRatio = calcPct(rs.Wins, rs.Wins + rs.Losses)
		}

		// A running tally of the counts we've seen so far
		overall.Games += s.Games
		overall.Wins += s.Wins
		overall.Losses += s.Losses

		summaries = append(summaries, &s)
	}

	overall.WinRatio = calcPct(overall.Wins, overall.Wins + overall.Losses)

	// Put the "overall" entry first in the list so it sorts in a stable fashion.
	summaries = append([]*GameTypeSummaryBase{&overall}, summaries...)

	sort.Sort(ByGames(summaries))

	return summaries, nil
}

// InfoBase is the view-agnostic representation of player information.
type InfoBase struct {
	PlayerID  int
	Nick      *models.MultiNick
	ActiveInd bool
	CreateDt  *models.MultiDt
}

// InfoData retrieves information about a given server.
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

	return &InfoBase{
		PlayerID:  rawPlayer.PlayerID,
		Nick:      nick,
		ActiveInd: rawPlayer.ActiveInd,
		CreateDt:  dt,
	}, nil
}
