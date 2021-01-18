package models

import (
	"database/sql"
	"time"
)

// scanPlayerSkillMatchResults is a helper function to parse rows into the information needed
// by the skill algorithm.
func scanPlayerSkillMatchResults(rows *sql.Rows) ([]*PlayerSkillMatchResult, error) {
	var results []*PlayerSkillMatchResult
	for rows.Next() {
		var s PlayerSkillMatchResult
		var durationSecs int
		var alivetimeMS int

		err := rows.Scan(&s.PlayerID, &s.GameID, &s.GameTypeCd, &durationSecs, &s.Score, &alivetimeMS, &s.Mu, &s.Sigma)

		s.Duration = time.Duration(durationSecs) * time.Second
		s.AliveTime = time.Duration(alivetimeMS) * time.Millisecond
		if err != nil {
			return nil, err
		}

		results = append(results, &s)
	}

	return results, nil
}

// RMatchResultsByGameID retrives match results suitable for input to the skill calculation algorithm.
func (ds *PGDatastore) RMatchResultsByGameID(gameID int) ([]*PlayerSkillMatchResult, error) {
	sql := `select pgs.player_id, g.game_id, g.game_type_cd, coalesce(cast(extract(epoch
		from g.duration) as integer)) as duration, pgs.score,
		cast(coalesce(extract(epoch from pgs.alivetime), 0)*1000 as integer) as alivetime, 
		ps.mu, ps.sigma
		from player_game_stats pgs join games g on pgs.game_id = g.game_id
		left outer join player_skills ps on g.game_type_cd = ps.game_type_cd
		where g.game_id = $1
		and pgs.player_id > 2
		order by pgs.scoreboardpos`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayerSkillMatchResults(rows)
}
