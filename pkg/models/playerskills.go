package models

import (
	"bytes"
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

// scanPlayerSkills scans the rows returned by a query to the player_skills table.
func scanPlayerSkills(rows *sql.Rows) ([]*PlayerSkill, error) {
	var results []*PlayerSkill
	for rows.Next() {
		var s PlayerSkill

		err := rows.Scan(&s.PlayerID, &s.GameTypeCd, &s.Mu, &s.Sigma, &s.ActiveInd, &s.CreateDt, &s.UpdateDt)
		if err != nil {
			return nil, err
		}

		results = append(results, &s)
	}

	return results, nil
}

// RPlayerSkills reads PlayerSkill values by PlayerID and optionally the game type.
func (ds *PGDatastore) RPlayerSkills(playerID int, gameTypeCd string) ([]*PlayerSkill, error) {
	var sqlBuf bytes.Buffer

	params := make([]interface{}, 0)

	sqlBuf.WriteString(`select player_id, game_type_cd, mu, sigma, active_ind, create_dt, update_dt
	from player_skills
	where player_id = $1 and active_ind = true `)

	params = append(params, playerID)

	if !(gameTypeCd == "" || gameTypeCd == "overall" || gameTypeCd == "all") {
		sqlBuf.WriteString("and gameTypeCd = $2")
		params = append(params, gameTypeCd)
	}

	sqlBuf.WriteString("order by Mu desc ")

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayerSkills(rows)

}

// CPlayerSkill inserts a PlayerSkill record into the database.
func (ds *PGDatastore) CPlayerSkill(tx *sql.Tx, ps PlayerSkill) error {
	sql := `insert into player_skills (player_id, game_type_cd, mu, sigma, active_ind) 
		values ($1, $2, $3, $4, $5) `

	_, err := tx.Exec(sql, ps.PlayerID, ps.GameTypeCd, ps.Mu, ps.Sigma, true)
	if err != nil {
		return err
	}

	return nil
}

// UPlayerSkill updates a PlayerSkill record in the database.
func (ds *PGDatastore) UPlayerSkill(tx *sql.Tx, ps PlayerSkill) error {
	sql := `update player_skills set mu=$1, sigma=$2, update_dt = now() at time zone 'UTC'
	where player_id = $3 and game_type_cd = $4`

	_, err := tx.Exec(sql, ps.Mu, ps.Sigma, ps.PlayerID, ps.GameTypeCd)
	if err != nil {
		return err
	}

	return nil
}
