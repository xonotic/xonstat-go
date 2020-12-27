package player

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// EloInfoBase is Elo information for a player for a single game type. 
// It's just an alias of the struct in the models package for now.
type EloInfoBase models.PlayerElo

// EloInfoData returns the view-agnostic elo data for a given player by ID.
func EloInfoData(db models.Datastore, hashkey string) ([]*EloInfoBase, error) {
	rawElos, err := db.RPlayerElosByHashkey(hashkey)
	if err != nil {
		return nil, err
	}

	// Cast model types to the base type.
	elos := make([]*EloInfoBase, len(rawElos))
	for i, e := range rawElos {
		eb := EloInfoBase(*e)
		elos[i] = &eb
	}

	return elos, nil 
}
