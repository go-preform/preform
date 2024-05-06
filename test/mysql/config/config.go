package config

import "fmt"

var (
	MysqlConnStr = fmt.Sprintf("%s:%s@/%s?parseTime=true", "root", "123456", "mysql")
	//MysqlConnStr = fmt.Sprintf("%s:%s@/%s?parseTime=true", "user", "password", "mysql")
)
