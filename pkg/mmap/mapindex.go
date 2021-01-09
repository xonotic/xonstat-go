package mmap

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// IndexData retrieves information about the list of servers.
func IndexData(db models.Datastore, start, limit int, nameFragment string) ([]*InfoBase, error) {
	rawMaps, err := db.RMapIndex(start, limit, nameFragment)
	if err != nil {
		return nil, err
	}

	maps := make([]*InfoBase, len(rawMaps))
	for i, m := range rawMaps {
		dt, err := models.NewMultiDt(m.CreateDt)
		if err != nil {
			return nil, err
		}

		ib := &InfoBase{
			MapID:  m.MapID,
			Name:      m.Name,
			CreateDt:  dt,
		}

		maps[i] = ib
	}

	return maps, nil
}
