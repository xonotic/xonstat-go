package models

import (
	"fmt"
	"strings"
)

// RSearchPlayer performs player searches by fragments of the nickname.
func (ds *PGDatastore) RSearchPlayer(nickFragment string) ([]*Player, error) {
	sql := `select p.player_id, p.nick, p.stripped_nick, p.location, p.email_addr, 
	p.active_ind, p.create_dt
	from players p 
	where UPPER(p.stripped_nick) like $1;`

	fragment := fmt.Sprintf("%%s%", strings.ToUpper(nickFragment))

	rows, err := ds.db.Query(sql, fragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]*Player, 0)
	for rows.Next() {
		var p Player
		err := rows.Scan(&p.PlayerID, &p.Nick, &p.StrippedNick, &p.Location, &p.EmailAddr, &p.ActiveInd, &p.CreateDt)
		if err != nil {
			return nil, err
		}

		players = append(players, &p)
	}

	return players, nil
}
