package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

func scanRecentGames(rows *sql.Rows) ([]*RecentGame, error) {
	var rgs []*RecentGame
	for rows.Next() {
		var rg RecentGame

		err := rows.Scan(&rg.GameID, &rg.GameTypeCd, &rg.WinningTeam, &rg.CreateDt, &rg.GameTypeDescr,
			&rg.ServerID, &rg.ServerName, &rg.MapID, &rg.MapName, &rg.WinningPlayerID, &rg.WinningNick,
			&rg.MatchID)

		if err != nil {
			return nil, err
		}

		rgs = append(rgs, &rg)
	}

	return rgs, nil
}

// RRecentGamesCutoff is an optimization of the RRecentGames query that uses
// a CTE for a large speedup. The speedup comes from the presence of a cutoff, 
// allowing the query to aggressively prune the records it looks at.
func (ds *PGDatastore) RRecentGamesCutoff(serverID int, mapID int, playerID int,
	gameTypeCd string, cutoff *time.Time, startGameID int,
	endGameID int, limit int) (string, []interface{}) {

	// Build up the SQL that will eventually be executed.
	// The filter conditions are built up separately.
	var sqlBuf bytes.Buffer
	var filterBuf bytes.Buffer

	// Keep track of the bind placeholders and their parameters
	placeholder := 1
	params := make([]interface{}, 0)

	sqlBuf.WriteString(`
	WITH player_game_stats_w AS (
		SELECT 
		   g.server_id, pgs.game_id, g.map_id, pgs.player_id, pgs.nick, 
		   rank() OVER (partition by g.game_id order by pgs.scoreboardpos ASC) AS scoreboardpos, 
		   g.game_type_cd, g.winner, g.create_dt, g.match_id
		FROM 
			player_game_stats pgs join games g using (game_id)
		WHERE 
			%s
	)
	SELECT 
		pgsw.game_id, pgsw.game_type_cd, pgsw.winner, pgsw.create_dt, cdg.descr, 
		pgsw.server_id, s.name, pgsw.map_id, m.name, pgsw.player_id, pgsw.nick, pgsq.match_id
	FROM 
		player_game_stats_w pgsw
		JOIN maps m ON (pgsw.map_id = m.map_id)
		JOIN cd_game_type cdg ON (pgsw.game_type_cd = cdg.game_type_cd)
		JOIN servers s ON (pgsw.server_id = s.server_id)
	WHERE 
		pgsw.scoreboardpos = 1
	ORDER BY 
		pgsw.create_dt DESC
	LIMIT $1;`)

	placeholder++
	params = append(params, limit)

	// A cutoff is present, so add a useful time bound that greatly limits the numer of rows searched.
	filterBuf.WriteString(
		fmt.Sprintf("g.create_dt between $%d and $%d ", placeholder, placeholder+1),
	)

	// Note that these are the same placeholder values for cutoff used on a different table
	filterBuf.WriteString(
		fmt.Sprintf("and pgs.create_dt between $%d and $%d ", placeholder, placeholder+1),
	)

	placeholder += 2
	params = append(params, cutoff)
	params = append(params, time.Now().UTC())

	if serverID != -1 {
		filterBuf.WriteString(fmt.Sprintf("and g.server_id = $%d ", placeholder))
		placeholder++
		params = append(params, serverID)
	}

	if mapID != -1 {
		filterBuf.WriteString(fmt.Sprintf("and g.map_id = $%d ", placeholder))
		placeholder++
		params = append(params, mapID)
	}

	if playerID != -1 {
		// Constrain the list of games returned to those that contained that player.
		// Unfortunately this can't be parameterized/bound.
		filterBuf.WriteString(fmt.Sprintf("and g.players @> ARRAY[%d] ", playerID))
	} 

	if gameTypeCd != "" {
		filterBuf.WriteString(fmt.Sprintf("and g.game_type_cd = $%d ", placeholder))
		placeholder++
		params = append(params, gameTypeCd)
	}

	// Useful for pagination in different ways.
	if startGameID != -1 {
		filterBuf.WriteString(fmt.Sprintf("and g.game_id <= $%d ", placeholder))
		placeholder++
		params = append(params, startGameID)
	}

	if endGameID != -1 {
		filterBuf.WriteString(fmt.Sprintf("and g.game_id >= $%d ", placeholder))
		placeholder++
		params = append(params, endGameID)
	}

	// We'll have the SQL with a placeholder for the filters, then a string for the filters 
	// themselves. Combine them to get the full SQL to be executed.
	sqlWithoutFilters := sqlBuf.String()
	filters := filterBuf.String()

	return  fmt.Sprintf(sqlWithoutFilters, filters), params
}

// RRecentGamesUnbounded constructs SQL and corresponding params for retrieving 
// recent games according to filter criteria. For the ID values, pass -1 to 
// exclude them from the query.
func (ds *PGDatastore) RRecentGamesUnbounded(serverID int, mapID int, playerID int,
	gameTypeCd string, cutoff *time.Time, startGameID int,
	endGameID int, limit int) (string, []interface{}) {

	// Build up the SQL that will eventually be executed.
	var sqlBuf bytes.Buffer

	// Keep track of the bind placeholders and their parameters
	placeholder := 1
	params := make([]interface{}, 0)

	sqlBuf.WriteString(`
	SELECT 
	    g.game_id, g.game_type_cd, g.winner, 
	    g.create_dt, cdg.descr, s.server_id, s.name, m.map_id, m.name, 
	    pgs.player_id, pgs.nick, g.match_id
	FROM 
	    games g, servers s, maps m, player_game_stats pgs, cd_game_type cdg
	WHERE 
	    g.server_id = s.server_id
	    and g.map_id = m.map_id
	    and g.game_type_cd = cdg.game_type_cd
	    and g.game_id = pgs.game_id 
	    and pgs.scoreboardpos = (select min(scoreboardpos) from player_game_stats where game_id = g.game_id) `)

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
	} 

	if gameTypeCd != "" {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_type_cd = $%d ", placeholder))
		placeholder++
		params = append(params, gameTypeCd)
	}

	// Useful for pagination in different ways.
	if startGameID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_id <= $%d ", placeholder))
		placeholder++
		params = append(params, startGameID)
	}

	if endGameID != -1 {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_id >= $%d ", placeholder))
		placeholder++
		params = append(params, endGameID)
	}

	sqlBuf.WriteString("order by g.create_dt desc ")
	sqlBuf.WriteString(fmt.Sprintf("limit $%d ", placeholder))
	placeholder++
	params = append(params, limit)

	return sqlBuf.String(), params
}

// RRecentGames retrieves recent games according to the filter criteria.
// For the ID values, pass -1 to exclude them from the query.
func (ds *PGDatastore) RRecentGames(serverID int, mapID int, playerID int,
	gameTypeCd string, cutoff *time.Time, startGameID int,
	endGameID int, limit int) ([]*RecentGame, error) {

	var sql string
	var params []interface{}

	if cutoff != nil {
		// If we have a cutoff value, use the faster version!
		sql, params = ds.RRecentGamesCutoff(serverID, mapID, playerID, gameTypeCd, 
			cutoff, startGameID, endGameID, limit)
	} else {
		// Otherwise we'll use the unbounded version as a default.
		sql, params = ds.RRecentGamesUnbounded(serverID, mapID, playerID, gameTypeCd, 
			cutoff, startGameID, endGameID, limit)
	}

	rows, err := ds.db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRecentGames(rows)
}

// IsWeaponInfoGameType determines if retrieving weapon info makes sense for the given game type.
func IsWeaponInfoGameType(gameTypeCd string) bool {
	switch gameTypeCd {
	case "", "overall":
		return true
	case "as", "ca", "ctf", "dm", "dom", "duel", "ft", "freezetag", "ka":
		return true
	case "keepaway", "kh", "rune", "tdm":
		return true
	default:
		return false
	}
}

// RGameIDsByPlayerID retrieves the recent N game IDs for further queries.
func (ds *PGDatastore) RGameIDsByPlayerID(playerID, limit int, gameTypeCd string) ([]int, error) {
	if playerID <= 2 {
		return nil, fmt.Errorf("invalid player ID")
	}

	// Some game types don't make sense with respect to weapon data...
	if !IsWeaponInfoGameType(gameTypeCd) {
		return nil, fmt.Errorf("unsupported game type '%s'", gameTypeCd)
	}

	// Clamp the limit of data retrieved
	if limit < 20 {
		limit = 20
	}

	if limit > 50 {
		limit = 50
	}

	// Keep track of the bind placeholders and their parameters
	placeholder := 1
	params := make([]interface{}, 0)

	// Build up the SQL that will eventually be executed.
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString(fmt.Sprintf("select g.game_id from games g where g.players @> ARRAY[%d] ", playerID))

	if gameTypeCd != "" && gameTypeCd != "overall" {
		sqlBuf.WriteString(fmt.Sprintf("and g.game_type_cd = $%d ", placeholder))
		placeholder++
		params = append(params, gameTypeCd)
	}

	// Clamp the data to the past year for performance reasons.
	sqlBuf.WriteString("and g.create_dt > (now() at time zone 'utc' - interval '1 year') ")

	sqlBuf.WriteString("order by g.game_id desc ")

	sqlBuf.WriteString(fmt.Sprintf("limit $%d ", placeholder))
	placeholder++
	params = append(params, limit)

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []int
	for rows.Next() {
		var gameID int

		err := rows.Scan(&gameID)

		if err != nil {
			return nil, err
		}

		gameIDs = append(gameIDs, gameID)
	}

	return gameIDs, nil
}