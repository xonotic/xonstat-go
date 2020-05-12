package models

// RServerByHashkey retrives a server record by the server's hashkey value
func (ds *PGDatastore) RServerByHashkey(hashkey string) (*Server, error) {
	sql := `select server_id, name, location, ip_addr, port, hashkey, public_key, revision, 
	pure_ind, impure_cvars, elo_ind, active_ind, create_dt
	from servers
	where hashkey = $1`

	rows, err := ds.db.Query(sql, hashkey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var s Server
	for rows.Next() {
		err := rows.Scan(&s.ServerId, &s.Name, &s.Location, &s.IPAddr, &s.Port, &s.HashKey, &s.PublicKey,
			&s.Revision, &s.PureInd, &s.ImpureCvars, &s.EloInd, &s.ActiveInd, &s.CreateDt)
		if err != nil {
			return nil, err
		}
	}

	return &s, nil
}
