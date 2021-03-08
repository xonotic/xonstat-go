package server

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// IndexData retrieves information about the list of servers.
func IndexData(db models.Datastore, start, limit int, nameFragment string) ([]*InfoBase, error) {
	rawServers, err := db.RServerIndex(start, limit, nameFragment)
	if err != nil {
		return nil, err
	}

	servers := make([]*InfoBase, len(rawServers))
	for i, s := range rawServers {
		name := models.NewMultiNick(s.Name.String)
		dt, err := models.NewMultiDt(s.CreateDt)
		if err != nil {
			return nil, err
		}

		ib := &InfoBase{
			ServerID:  s.ServerID,
			Name:      s.Name.String,
			NameHTML:  name.NickHTML,
			IPAddr:    s.IPAddr.String,
			Port:      int(s.Port.Int64),
			Revision:  s.Revision.String,
			ActiveInd: s.ActiveInd,
			CreateDt:  dt,
		}

		servers[i] = ib
	}

	return servers, nil
}
