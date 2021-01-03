package player

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// IndexData retrieves information about the list of players.
func IndexData(db models.Datastore, start, limit int, nickFragment string) ([]*InfoBase, error) {
	rawPlayers, err := db.RPlayerIndex(start, limit, nickFragment)
	if err != nil {
		return nil, err
	}

	players := make([]*InfoBase, len(rawPlayers))
	for i, p := range rawPlayers {
		nick := models.NewMultiNick(p.Nick.String)
		dt, err := models.NewMultiDt(p.CreateDt)
		if err != nil {
			return nil, err
		}

		ib := &InfoBase{
			PlayerID:  p.PlayerID,
			Nick:      nick,
			ActiveInd: p.ActiveInd,
			CreateDt:  dt,
		}

		players[i] = ib
	}

	return players, nil
}
