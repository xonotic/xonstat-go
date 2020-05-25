package models

import (
	"database/sql"
	"fmt"
	"strings"
)

// CPlayer inserts a Player record into the database.
func (ds *PGDatastore) CPlayer(tx *sql.Tx, player Player) (int64, error) {
	var pid int64

	seqVal, err := ds.nextSeqVal("players_player_id_seq")
	if err != nil {
		return pid, err
	}
	pid = seqVal

	if player.Nick.Valid && player.Nick.String == "Anonymous Player" {
		player.Nick = sql.NullString {
			Valid: true, 
			String: fmt.Sprintf("Anonymous Player #%d", pid),
		}

		player.StrippedNick = sql.NullString {
			Valid: true, 
			String: fmt.Sprintf("Anonymous Player #%d", pid),
		}
	}

	isql := `insert into players (player_id, nick, stripped_nick, location, email_addr) 
	values ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(isql, pid, player.Nick, player.StrippedNick, player.Location, 
		player.EmailAddr)

    if err != nil {
		return pid, err
	}

	return pid, nil
}

// RPlayersByHashkeyMulti finds multiple players by their hashkeys, returning back a map
// indexed by the hashkey with player record pointers as values.
func (ds *PGDatastore) RPlayersByHashkeyMulti(hashkeys []string) (map[string]*Player, error) {
	var quotedHashkeys []string
	for _, hk := range hashkeys {
		quotedHashkeys = append(quotedHashkeys, fmt.Sprintf("'%s'", hk))
	}
	hashKeyInStr := strings.Join(quotedHashkeys, ", ")

	sql := fmt.Sprintf(`select p.player_id, p.nick, p.stripped_nick, p.location, p.email_addr, 
	p.active_ind, p.create_dt, ph.hashkey
	from players p join hashkeys ph on p.player_id = ph.player_id
	where ph.hashkey in (%s)`, hashKeyInStr)

	rows, err := ds.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players map[string]*Player
	for rows.Next() {
		var p Player
		var hashkey string

		err := rows.Scan(&p.PlayerID, &p.Nick, &p.StrippedNick, &p.Location, &p.EmailAddr, 
			&p.ActiveInd, &p.CreateDt, &hashkey)

		if err != nil {
			return nil, err
		}

		players[hashkey] = &p
	}

	return players, nil
}

// UPlayer updates a Player record in the database.
func (ds *PGDatastore) UPlayer(tx *sql.Tx, player Player) error {
	sql := `update players set nick=$1, stripped_nick=$2
	where player_id = $3`

	_, err := tx.Exec(sql, player.Nick, player.StrippedNick, player.PlayerID)

	if err != nil {
		return err
	}

	return nil
}