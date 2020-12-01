package player

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// InfoBase is the view-agnostic representation of player information.
type InfoBase struct {
	PlayerID int
	Nick *models.MultiNick
	ActiveInd bool
	CreateDt *models.MultiDt
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

	return &InfoBase {
		PlayerID: rawPlayer.PlayerID,
		Nick: nick,
		ActiveInd: rawPlayer.ActiveInd,
		CreateDt: dt,
	}, nil
}