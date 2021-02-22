package skill

import (
	"time"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// InfoBase is the view-agnostic representation of a player_skill record.
type InfoBase struct {
	PlayerID   int
	GameTypeCd string
	Mu         float64
	Sigma      float64
	ActiveInd  bool
	CreateDt   time.Time
	UpdateDt   time.Time
}

// InfoData retrieves information about a given server.
func InfoData(db models.Datastore, playerID int, gameTypeCd string) ([]*InfoBase, error) {
	rawSkills, err := db.RPlayerSkills(playerID, gameTypeCd)
	if err != nil {
		return nil, err
	}

	skills := make([]*InfoBase, len(rawSkills))
	for i, v := range rawSkills {
		skills[i] = &InfoBase{
			PlayerID: v.PlayerID,
			GameTypeCd: v.GameTypeCd,
			Mu: v.Mu,
			Sigma: v.Sigma,
			ActiveInd: v.ActiveInd,
			CreateDt: v.CreateDt,
			UpdateDt: v.UpdateDt,
		}
	}

	return skills, nil
}