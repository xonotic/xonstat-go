package models

import (
	"bytes"
	"fmt"
	"time"
)

// RRecentGames retrieves recent games according to the filter criteria.
// For the ID values, pass -1 to exclude them from the query.
func (ds *PGDatastore) RRecentGames(serverID int, mapID int, playerID int,
	gameTypeCd string, cutoff time.Time, forcePlayerID bool, startGameID int,
	endGameID int, limit int) ([]*RecentGame, error) {

	// Build up the SQL that will eventually be executed.
	var sqlBuf bytes.Buffer

	// Keep track of the bind placeholders and their parameters
	placeholder := 1
	params := make([]interface{}, 0)

	sqlBuf.WriteString(`select g.game_id, g.game_type_cd, g.winner, 
	g.create_dt, cdg.descr, s.server_id, s.name, m.map_id, m.name, 
	pgs.player_id, pgs.nick, pgs.rank, pgs.team, pgs.elo_delta
	from games g, servers s, maps m, player_game_stats pgs, cd_game_type cdg
	where g.server_id = s.server_id
	and g.map_id = m.map_id
	and g.game_type_cd = cdg.game_type_cd
	and g.game_id = pgs.game_id `)

	if serverID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and s.server_id = $%d ", placeholder))
		placeholder++
		params = append(params, serverID)
	}

	if mapID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and m.map_id = $%d ", placeholder))
		placeholder++
		params = append(params, mapID)
	}

	if playerID != -1 {
		// Constrain the list of games returned to those that contained that player.
		// Unfortunately this can't be parameterized/bound.
		sqlBuf.WriteString(fmt.Sprintf("and g.players @> ARRAY[%d] ", playerID))

		if forcePlayerID {
			// Forcing the player_id means that the pgstat row returned is from that player and
			// not necessarily the winner (the default).
			sqlBuf.WriteString(fmt.Sprintf("and pgs.player_id = $%d ", placeholder))
			placeholder++
			params = append(params, playerID)
		}
	} else {
		// If we're NOT looking for a specific player_id, return the pgstat record of the winner.
		sqlBuf.WriteString("and pgs.scoreboardpos = 1 ")
	}

	if gameTypeCd != "" {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_type_cd = $%d ", placeholder))
		placeholder++
		params = append(params, gameTypeCd)
	}

	// Useful for pagination in different ways.
	if startGameID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_id >= $%d ", placeholder))
		placeholder++
		params = append(params, startGameID)
	}

	if endGameID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_id <= $%d ", placeholder))
		placeholder++
		params = append(params, endGameID)
	}

	// This section adds a useful time bound that greatly limits the numer of rows searched.
	sqlBuf.WriteString(
		fmt.Sprintf("and g.create_dt between $%d and $%d", placeholder, placeholder+1),
	)

	placeholder += 2
	params = append(params, cutoff)
	params = append(params, time.Now().UTC)

	sqlBuf.WriteString("order by g.create_dt desc ")
	sqlBuf.WriteString(fmt.Sprintf("limit $%d ", placeholder))
	placeholder++
	params = append(params, limit)

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rgs []*RecentGame
	for rows.Next() {
		var rg RecentGame

		err := rows.Scan(&rg.GameID, &rg.GameTypeCd, &rg.Winner, &rg.StartDt, &rg.GameTypeDescr,
			&rg.ServerID, &rg.ServerName, &rg.MapID, &rg.MapName, &rg.PlayerID, &rg.Nick, &rg.Rank,
			&rg.Team, &rg.EloDelta)

		if err != nil {
			return nil, err
		}

		rgs = append(rgs, &rg)
	}

	return rgs, nil
}
