package models

import (
	"database/sql"
	"fmt"
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
	order by (case when active_ind then 0 else 1 end), hashkey, create_dt
	`

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

// RServersByIPAndPort retrieves server records by their IP address and port.
func (ds *PGDatastore) RServersByIPAndPort(ipAddr string, port int) ([]*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where ip_addr = $1 and port = $2
	order by create_dt`

	rows, err := ds.db.Query(sql, ipAddr, port)
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
