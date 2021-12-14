package postgresctl

import (
	"errors"
	"fmt"

	"database/sql"

	_ "github.com/lib/pq" // for postgresql
)

// Connect connects to postgresql database
func (db *PostgresDB) Connect() (err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=3", db.Cfg.Dbhost, db.Cfg.Dbport, db.Cfg.Dbuser, db.Cfg.Dbpassword, db.Cfg.Dbname)
	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	err = conn.Ping()
	if err != nil {
		return err
	}

	db.Conns = append(db.Conns, conn)

	return err
}

func (db *PostgresDB) GetDbHost() string {
	return db.Cfg.Dbhost
}

func (db *PostgresDB) GetNbConn() int {
	return len(db.Conns)
}

// Close closes connection to PostgresDB
func (db *PostgresDB) Close() {
	for idx, conn := range db.Conns {
		fmt.Printf("Close cnx id %d\n", idx)
		conn.Close()
	}
}

// CheckConn check the connections and returns true if ok and false if connection is dead
func (db *PostgresDB) CheckConn() bool {
	if len(db.Conns) == 0 {
		return false
	}
	err := db.Conns[0].Ping()
	return err == nil
}

func (db *PostgresDB) CalcDatabaseSize() (err error) {
	var rows *sql.Rows

	if db.GetNbConn() == 0 {
		return errors.New("Not connected to DB")
	}
	rows, err = db.Conns[0].Query("SELECT pg_database_size('" + db.Cfg.Dbname + "') ")

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&db.Size)
		if err != nil {
			return err
		}
	}
	// get any error encountered during iteration
	err = rows.Err()
	return err
}

func (db *PostgresDB) GetDBSizeGo() int {
	return db.Size / 1024 / 1024 / 1024
}

func (db *PostgresDB) CalcCnx() (err error) {
	var rows *sql.Rows
	if db.GetNbConn() == 0 {
		return errors.New("Not connected to DB")
	}
	rows, err = db.Conns[0].Query(`select max_conn,used,res_for_super,max_conn-used-res_for_super res_for_normal 
	from 
	  (select count(*) used from pg_stat_activity) t1,
	  (select setting::int res_for_super from pg_settings where name=$$superuser_reserved_connections$$) t2,
	  (select setting::int max_conn from pg_settings where name=$$max_connections$$) t3;`)

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&db.NbMaxConnections, &db.NbUsedConnections, &db.NbReservedForSuperUser, &db.NbReservedForNormalUser)
		if err != nil {
			return err
		}
	}
	// get any error encountered during iteration
	err = rows.Err()
	return err
}

// CollectInfos collects informations about sensors
func (db *PostgresDB) CollectInfos() (err error) {
	if !db.CheckConn() {
		fmt.Println("CollectInfos : not connected to " + db.Cfg.Dbhost)
		db.Connect()
		if !db.CheckConn() {
			return err
		}
	}
	err = db.CalcDatabaseSize()
	if err != nil {
		return err
	}
	err = db.CalcCnx()

	return err
}
