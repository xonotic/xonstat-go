package models

import (
	"bytes"
	"fmt"
	"strings"
)

// BlankStart is a blank or empty starting point value.
const BlankStart = -1

// RPlayerIndex performs player searches by simple pagination.
func (ds *PGDatastore) RPlayerIndex(start, limit int, nickFragment string) ([]*Player, error) {
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString(`select p.player_id, p.nick, p.stripped_nick, p.location, p.email_addr, 
	p.active_ind, p.create_dt
	from players p 
	where p.nick not like 'Anonymous Player%'
	and p.player_id > 2`)

	// We might not have a start value, so we have to keep track of the bind params
	// and their placeholder numbers.
	params := make([]interface{}, 0)
	paramIndex := 1
	if start != BlankStart {
		sqlBuf.WriteString(fmt.Sprintf("and player_id < $%d ", paramIndex))
		params = append(params, start)
		paramIndex++
	}

	if nickFragment != "" {
		nickFragment = fmt.Sprintf("%%%s%%", strings.ToUpper(nickFragment))
		sqlBuf.WriteString(fmt.Sprintf("and UPPER(p.stripped_nick) like $%d ", paramIndex))
		params = append(params, nickFragment)
		paramIndex++
	}

	sqlBuf.WriteString(fmt.Sprintf("order by p.player_id desc limit $%d;", paramIndex))
	params = append(params, limit)
	paramIndex++

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlayers(rows)
}

// RServerIndex performs server searches by simple pagination.
func (ds *PGDatastore) RServerIndex(start, limit int, nameFragment string) ([]*Server, error) {
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString(`select server_id, name, location, ip_addr, port, hashkey, 
	public_key, revision, pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where active_ind = true `)

	// We might not have a start value, so we have to keep track of the bind params
	// and their placeholder numbers.
	params := make([]interface{}, 0)
	paramIndex := 1
	if start != BlankStart {
		sqlBuf.WriteString(fmt.Sprintf("and server_id < $%d ", paramIndex))
		params = append(params, start)
		paramIndex++
	}

	if nameFragment != "" {
		nameFragment = fmt.Sprintf("%%%s%%", strings.ToUpper(nameFragment))
		sqlBuf.WriteString(fmt.Sprintf("and UPPER(name) like $%d ", paramIndex))
		params = append(params, nameFragment)
		paramIndex++
	}

	sqlBuf.WriteString(fmt.Sprintf("order by server_id desc limit $%d;", paramIndex))
	params = append(params, limit)
	paramIndex++

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}
