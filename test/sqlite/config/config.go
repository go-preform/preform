package config

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func InitDb(wd string) *sql.DB {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/preform_test_a.db", wd))
	if err != nil {
		panic(err)
	}
	return db
}
