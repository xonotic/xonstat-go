package models

import (
	"database/sql"
)

// scanServers is a helper function to parse full Server records out of a resultset.
func scanServers(rows *sql.Rows) ([]*Server, error) {
	var servers []*Server
	for rows.Next() {
		var s Server
		err := rows.Scan(&s.ServerId, &s.Name, &s.Location, &s.IPAddr, &s.Port, &s.HashKey, &s.PublicKey,
			&s.Revision, &s.PureInd, &s.ImpureCvars, &s.EloInd, &s.ActiveInd, &s.CreateDt)

		if err != nil {
			return nil, err
		}

		servers = append(servers, &s)
	}

	return servers, nil
}

// RServersByHashkey retrives server records by their hashkey value.
func (ds *PGDatastore) RServersByHashkey(hashkey string) ([]*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where hashkey = $1
	and active_ind = true
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
	and active_ind = true
	order by create_dt`

	rows, err := ds.db.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanServers(rows)
}

// CServer inserts a Server record into the database.
func (ds *PGDatastore) CServer(server Server) (int64, error) {
	sql := `insert into servers (name, location, ip_addr, port, hashkey, public_key, revision, 
		pure_ind, impure_cvars, elo_ind, active_ind)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning server_id`

	row := ds.db.QueryRow(sql, server.Name, server.Location, server.IPAddr, server.Port,
		server.HashKey, server.PublicKey, server.Revision, server.PureInd, server.ImpureCvars,
		true, true)

	var serverID int64
	err := row.Scan(&serverID)
	if err != nil {
		return serverID, err
	}

	return serverID, nil
}
