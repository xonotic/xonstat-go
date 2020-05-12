package models

import (
	"database/sql"
)

// scanServer is a helper function to parse full Server records out of already-fetched rows.
func scanServer(row *sql.Row) (*Server, error) {
	var s Server
	err := row.Scan(&s.ServerId, &s.Name, &s.Location, &s.IPAddr, &s.Port, &s.HashKey, &s.PublicKey,
		&s.Revision, &s.PureInd, &s.ImpureCvars, &s.EloInd, &s.ActiveInd, &s.CreateDt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// RServerByHashkey retrives a server record by the server's hashkey value
func (ds *PGDatastore) RServerByHashkey(hashkey string) (*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where hashkey = $1
	and active_ind = true
	order by hashkey, create_dt`

	row := ds.db.QueryRow(sql, hashkey)
	return scanServer(row)
}

// RServerByName retrives a server record by the server's name value
func (ds *PGDatastore) RServerByName(name string) (*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where name = $1
	and active_ind = true
	order by create_dt`

	row := ds.db.QueryRow(sql, name)
	return scanServer(row)
}
