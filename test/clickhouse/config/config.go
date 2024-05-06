package config

import (
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"time"
)

func ChConn() *sql.DB {
	chConn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
			//"flatten_nested":     0, driver not supported
		},
		DialTimeout: time.Second * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug:                true,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})
	chConn.SetMaxIdleConns(5)
	chConn.SetMaxOpenConns(10)
	chConn.SetConnMaxLifetime(time.Hour)
	return chConn
}
