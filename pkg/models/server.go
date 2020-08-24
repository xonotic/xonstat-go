package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CServer inserts a Server record into the database.
func (ds *PGDatastore) CServer(tx *sql.Tx, server Server) (int64, error) {
	sql := `insert into servers (name, location, ip_addr, port, hashkey, public_key, revision, 
		pure_ind, impure_cvars, elo_ind, active_ind)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning server_id`

	row := tx.QueryRow(sql, server.Name, server.Location, server.IPAddr, server.Port,
		server.HashKey, server.PublicKey, server.Revision, server.PureInd, server.ImpureCvars,
		true, true)

	var serverID int64
	err := row.Scan(&serverID)
	if err != nil {
		return serverID, err
	}

	return serverID, nil
}

// scanServers is a helper function to parse full Server records out of a resultset.
func scanServers(rows *sql.Rows) ([]*Server, error) {
	var servers []*Server
	for rows.Next() {
		var s Server
		err := rows.Scan(&s.ServerID, &s.Name, &s.Location, &s.IPAddr, &s.Port, &s.HashKey, &s.PublicKey,
			&s.Revision, &s.PureInd, &s.ImpureCvars, &s.EloInd, &s.ActiveInd, &s.CreateDt)

		if err != nil {
			return nil, err
		}

		servers = append(servers, &s)
	}

	return servers, nil
}

// RServerByID retrives a server record by its ID value.
func (ds *PGDatastore) RServerByID(ID int) (*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where server_id = $1`

	rows, err := ds.db.Query(sql, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers, err := scanServers(rows)
	if len(servers) != 1 {
		return nil, fmt.Errorf("Unable to retrieve just one server")
	}

	return servers[0], nil
}

// RServersByHashkey retrives server records by their hashkey value.
func (ds *PGDatastore) RServersByHashkey(hashkey string) ([]*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where hashkey = $1
	order by hashkey, create_dt`

	rows, err := ds.db.Query(sql, hashkey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanServers(rows)
}

// RServersByName retrives server records by their name value.
func (ds *PGDatastore) RServersByName(name string) ([]*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where name = $1
	order by create_dt`

	rows, err := ds.db.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServers(rows)
}

// UServer updates a Server record in the database.
func (ds *PGDatastore) UServer(tx *sql.Tx, server Server) error {
	sql := `update servers set name=$1, location=$2, ip_addr=$3, port=$4, hashkey=$5, 
	public_key=$6, revision=$7, pure_ind=$8, impure_cvars=$9 
	where server_id = $10`

	_, err := tx.Exec(sql, server.Name, server.Location, server.IPAddr, server.Port,
		server.HashKey, server.PublicKey, server.Revision, server.PureInd, server.ImpureCvars,
		server.ServerID)

	if err != nil {
		return err
	}

	return nil
}

// RServerTopMaps finds the most active maps played on a server over a given time period.
func (ds *PGDatastore) RServerTopMaps(serverID int, cutoff *time.Time, limit int) ([]*ActiveMap, error) {
	sql := `SELECT row_number() OVER (ORDER BY count(*) DESC) AS rank, 
	games.map_id AS games_map_id, maps.name AS maps_name, count(*) AS times_played
	FROM games, maps
	WHERE maps.map_id = games.map_id 
	AND games.server_id = $1 
	AND games.create_dt > $2 
	GROUP BY games.map_id, maps.name 
	ORDER BY count(*) DESC
	LIMIT $3`

	rows, err := ds.db.Query(sql, serverID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topMaps []*ActiveMap
	for rows.Next() {
		var tm ActiveMap
		err := rows.Scan(&tm.SortOrder, &tm.MapID, &tm.MapName, &tm.Games)

		if err != nil {
			return nil, err
		}

		topMaps = append(topMaps, &tm)
	}

	return topMaps, nil
}