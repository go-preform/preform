package main

import (
	"database/sql"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/test/clickhouse/config"
	"os"
	"os/exec"
)

func main() {
	chConn := config.ChConn()
	prepareSchema(chConn)
	preformBuilder.BuildModel(chConn, "mainModel", "mainModel", "preform_test_a", "preform_test_b")

	fmt.Println("go test----------------------------")
	d, _ := os.Getwd()
	p := fmt.Sprintf("%s/model_test", d)
	cmd := exec.Command("go", "test", p)
	cmd.Dir = p
	out, err := cmd.CombinedOutput()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println(string(out), err)
}

func prepareSchema(myConn *sql.DB) {
	_, err := myConn.Exec("DROP DATABASE IF EXISTS preform_test_a")
	if err != nil {
		panic(err)
	}
	_, err = myConn.Exec("DROP DATABASE IF EXISTS preform_test_b")
	if err != nil {
		panic(err)
	}
	_, err = myConn.Exec(`CREATE DATABASE preform_test_a;`)
	if err != nil {
		panic(err)
	}
	_, err = myConn.Exec(`CREATE DATABASE preform_test_b;`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.user (
	id Int32,
	` + "`name`" + ` String,
	manager_ids Array(Int32) comment 'fk:preform_test_a.user.id',
	created_by Int32 comment 'fk:preform_test_a.user.id',
	created_at DateTime,
	logined_at Nullable(DateTime),
	detail Nested (age Int32, date_of_birth Date)
)
ENGINE = MergeTree
ORDER BY (id, created_at);`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.user_log (
	id UUID,
	user_id Int32 comment 'fk:preform_test_a.user.id',
	related_log_id Nullable(UUID) comment 'fk:preform_test_a.user_log.id',
	type Enum('Register' = 1, 'Login' = 2),
	detail Map(String, String)
)
ENGINE = MergeTree
ORDER BY (id, user_id);`)
	if err != nil {
		panic(err)
	}
}
