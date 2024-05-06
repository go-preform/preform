package config

import "fmt"

var (
	PgConnStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", 5432, "postgres", "123456", "postgres")
)
