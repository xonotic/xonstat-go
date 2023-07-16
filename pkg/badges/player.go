package badges

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	"github.com/antzucaro/qstr"
	_ "github.com/jackc/pgx"
)

// GameCount is the number of games played of the given game mode.
type GameCount struct {
	GameTypeCd string
	GameCount  int
}

// PlayerData holds aggregate statistics for players
type PlayerData struct {
	Nick         qstr.QStr
	StrippedNick string
	Kills        int
	Deaths       int
	Wins         int
	Losses       int
	PlayingTime  time.Duration
	GameCounts   []GameCount
}

// KDRatio returns the player'c Kill:Death ratio as a string
func (pd *PlayerData) KDRatio() float64 {
	if pd.Deaths > 0 {
		return float64(pd.Kills) / float64(pd.Deaths)
	} else {
		return 0.000
	}
}

// WinPct returns the player'c win percentage as a string
func (pd *PlayerData) WinPct() float64 {
	totalGames := pd.Wins + pd.Losses
	if totalGames > 0 {
		return float64(pd.Wins) / float64(totalGames) * 100
	} else {
		return 0.00
	}
}

// DurationString creates a human-readable duration string with a days component.
func DurationString(d time.Duration) string {
	minutes := uint64(d.Minutes())
	days := uint64(minutes / 1440)
	minutes -= days * 1440
	hours := uint64(minutes / 60)
	minutes -= hours * 60

	var buffer bytes.Buffer
	if days == 1 {
		buffer.WriteString("1 day")
	} else if days > 1 {
		buffer.WriteString(fmt.Sprintf("%d days", days))
	}

	if hours >= 1 && days >= 1 {
		buffer.WriteString(", ")
	}

	if hours == 1 {
		buffer.WriteString("1 hr")
	} else if hours > 1 {
		buffer.WriteString(fmt.Sprintf("%d hrs", hours))
	}

	if minutes >= 1 && hours >= 1 {
		buffer.WriteString(", ")
	}

	if minutes == 1 {
		buffer.WriteString("1 min")
	} else if minutes > 1 {
		buffer.WriteString(fmt.Sprintf("%d mins", minutes))
	}
	return buffer.String()
}

// PlayingTime constructs a human-readable duration string with a day component.
func (pd *PlayerData) PlayingTimeString() string {
	return DurationString(pd.PlayingTime)
}

// PlayerDataFetcher fetches player information from the database
type PlayerDataFetcher struct {
	db *sql.DB
}

// NewPlayerDataFetcher creates a new PlayerDataFetcher for obtaining
// player information from the database
func NewPlayerDataFetcher(connStr string) (*PlayerDataFetcher, error) {
	// establish a database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// connection pooling
	db.SetMaxIdleConns(5)

	pp := PlayerDataFetcher{db: db}
	return &pp, nil
}

// FindPlayers finds a list of player_id values according to certain criteria.
// If delta is set, it will look for players who have had activity in the last
// $delta hours. If limit is set, the total number of player_ids returned is
// limited to that amount.
func (pp *PlayerDataFetcher) FindPlayers(delta int, limit int) ([]int, error) {
	playersSQL := `SELECT distinct p.player_id 
	FROM players p JOIN player_agg_stats_mv pas on p.player_id = pas.player_id
    JOIN player_elos pe on p.player_id = pe.player_id
	WHERE p.active_ind = true
	AND p.player_id > 2
	AND p.nick IS NOT NULL`

	// constrain the time window if needed
	if delta > 0 {
		playersSQL += " AND pas.create_dt > now() - interval '" + fmt.Sprintf("%d", delta) + " hours'"
	}

	// limit the number of players if needed
	if limit > 0 {
		playersSQL += " LIMIT " + fmt.Sprintf("%d", limit)
	}

	// DEBUG
	// fmt.Println(playersSQL)

	rows, err := pp.db.Query(playersSQL)
	if err != nil {
		return nil, err
	}

	pids := make([]int, 0, 100)
	var pid int
	for rows.Next() {
		rows.Scan(&pid)
		pids = append(pids, pid)
	}

	return pids, nil
}

// genPlayerDataStmt generates the SQL statement string used to fetch
// the information used to populate PlayerData objects
func (pp *PlayerDataFetcher) genPlayerDataStmt(playerID int) string {
	query := `
SELECT
    p.nick,
    p.stripped_nick,
    upper(pa.game_type_cd) as game_type_cd,
    pa.games,
    pa.wins,
    pa.losses,
    pa.kills,
    pa.deaths,
    pa.alivetime
FROM
    player_agg_stats_mv pa
JOIN
    players p
        on p.player_id = pa.player_id
WHERE
   pa.player_id = %d
ORDER BY
   pa.games desc
;
`

	return fmt.Sprintf(query, playerID)
}

// GetPlayerData retrieves player information for the given player_id
func (pp *PlayerDataFetcher) GetPlayerData(playerID int) (*PlayerData, error) {
	sqlQuery := pp.genPlayerDataStmt(playerID)

	rows, err := pp.db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}

	pd := new(PlayerData)

	filled := false
	var nick, strippedNick, gameType string
	var games, wins, losses, kills, deaths, alivetime int
	var totalWins, totalLosses, totalKills, totalDeaths, totalAlivetime int
	gameCounts := make([]GameCount, 0)

	for rows.Next() {
		err := rows.Scan(&nick, &strippedNick, &gameType, &games, &wins, &losses, &kills, &deaths, &alivetime)
		if err != nil {
			panic(err)
		}

		gameCounts = append(gameCounts, GameCount{gameType, games})

		// did we fill in the player information yet?
		if !filled {
			pd.Nick = qstr.QStr(nick)
			pd.Nick = pd.Nick.Decode(qstr.XonoticDecodeKey)
			pd.StrippedNick = strippedNick
			filled = true
		}

		// DM, CTS, and KA do not count towards win percentage
		if gameType != "DM" && gameType != "CTS" && gameType != "KA" {
			totalWins += wins
			totalLosses += losses
		}

		totalKills += kills
		totalDeaths += deaths
		totalAlivetime += alivetime
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	pd.GameCounts = gameCounts
	pd.Kills = totalKills
	pd.Deaths = totalDeaths
	pd.Wins = totalWins
	pd.Losses = totalLosses
	pd.PlayingTime = time.Duration(totalAlivetime) * time.Minute

	return pd, nil
}
