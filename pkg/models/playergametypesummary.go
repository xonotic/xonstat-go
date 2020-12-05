package models

// RGameTypeSummariesByID retrieves summary information about a player's games.
func (ds *PGDatastore) RGameTypeSummariesByID(playerID int) ([]*PlayerGameTypeSummary, error) {
	sql := `
SELECT game_type_cd, 
	SUM(win) wins, 
	SUM(loss) losses 
FROM   (SELECT g.game_id, 
			g.game_type_cd, 
			CASE 
			  WHEN g.winner = pgs.team THEN 1 
			  WHEN pgs.scoreboardpos = 1 THEN 1 
			  ELSE 0 
			END win, 
			CASE 
			  WHEN g.winner = pgs.team THEN 0 
			  WHEN pgs.scoreboardpos = 1 THEN 0 
			  ELSE 1 
			END loss 
	 FROM   games g, 
			player_game_stats pgs 
	 WHERE  g.game_id = pgs.game_id 
	 AND pgs.player_id = $1 
	 AND g.players @> ARRAY[$1]) win_loss 
GROUP  BY game_type_cd `

	rows, err := ds.db.Query(sql, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*PlayerGameTypeSummary
	for rows.Next() {
		var s PlayerGameTypeSummary

		err := rows.Scan(&s.GameTypeCd, &s.Wins, &s.Losses)

		if err != nil {
			return nil, err
		}

		summaries = append(summaries, &s)
	}

	return summaries, nil
}
