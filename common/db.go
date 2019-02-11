package common

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Db struct {
	db *sql.DB
}

func (s *Db) Connect(host string, port int, user, pwd, database string) error {
	var err error
	s.db, err = sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", user, pwd, host, port, database))
	if err != nil {
		return err
	}
	s.db.SetMaxOpenConns(200)
	s.db.SetMaxIdleConns(100)
	return nil
}

func (s *Db) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(sql, args...)
}

func (s *Db) QueryRow(sql string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(sql, args...)
}

func (s *Db) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(sql, args...)
}
