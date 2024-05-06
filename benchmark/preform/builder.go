package main

import (
	"database/sql"
	"github.com/go-preform/preform/benchmark/config"
	"github.com/go-preform/preform/preformBuilder"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	pgxConn, _ := sql.Open("pgx", config.PgConnStr)
	preformBuilder.BuildModel(pgxConn, "preformModel", "preformModel", "preform_benchmark")
}
