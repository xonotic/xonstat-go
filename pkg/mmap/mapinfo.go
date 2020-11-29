package mmap

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// InfoBase is the view-agnostic information about a map.
type InfoBase struct {
	MapID    int
	Name     string
	CreateDt *models.MultiDt
}

// InfoData retrieves all of the relevant information about a map and returns it.
func InfoData(db models.Datastore, mapID int) (*InfoBase, error) {
	m, err := db.RMapByID(mapID)
	if err != nil {
		return nil, err
	}

	dt, err := models.NewMultiDt(m.CreateDt)
	if err != nil {
		return nil, err
	}

	return &InfoBase{
		MapID:   m.MapID,
		Name:    m.Name,
		CreateDt: dt,
	}, nil
}