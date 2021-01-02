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
	where UPPER(p.stripped_nick) LIKE $1;`

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

// RSearchServer performs server searches by fragments of their names
func (ds *PGDatastore) RSearchServer(nameFragment string) ([]*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where UPPER(name) LIKE $1;`

	fragment := fmt.Sprintf("%%s%", strings.ToUpper(nameFragment))

	rows, err := ds.db.Query(sql, fragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}

// RSearchMap performs map searches by fragments of their names
func (ds *PGDatastore) RSearchMap(nameFragment string) ([]*Map, error) {
	sql := `select map_id, name, version, pk3_name, curl_url, create_dt 
	from maps
	where UPPER(name) LIKE $1;`

	fragment := fmt.Sprintf("%%s%", strings.ToUpper(nameFragment))

	rows, err := ds.db.Query(sql, fragment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMaps(rows)
}
