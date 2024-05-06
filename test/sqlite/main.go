package main

import (
	"database/sql"
	"fmt"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/test/sqlite/config"
	"os"
	"os/exec"
)

func main() {
	wd, _ := os.Getwd()
	err := os.Remove(wd + "/preform_test_a.db")
	fmt.Println(err)
	err = os.Remove(wd + "/preform_test_b.db")
	fmt.Println(err)
	conn := config.InitDb(wd)
	prepareSchema(conn)
	preformBuilder.BuildModel(conn, "mainModel", "mainModel", "preform_test_a", "preform_test_b")

	fmt.Println("go test----------------------------")
	d, _ := os.Getwd()
	p := fmt.Sprintf("%s/model_test", d)
	cmd := exec.Command("go", "test", p, "-root="+d)
	cmd.Dir = d
	out, err := cmd.CombinedOutput()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println(string(out), err)
}

func prepareSchema(conn *sql.DB) {
	wd, _ := os.Getwd()
	_, err := conn.Exec(fmt.Sprintf(`
attach database '%s/preform_test_a.db' as 'preform_test_a';
attach database '%s/preform_test_b.db' as preform_test_b;

CREATE TABLE preform_test_a."user" (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	"name" TEXT NOT NULL,
	created_by INTEGER NOT NULL,
	created_at INTEGER NOT NULL, --type:datetime
	logined_at INTEGER NULL, --type:datetime
	detail jsonb DEFAULT NULL,
	config jsonb NOT NULL,
	extra_config jsonb NOT NULL,
	FOREIGN KEY(created_by) REFERENCES "user"(id)
);

CREATE TABLE preform_test_a.user_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	user_id INTEGER NOT NULL,
	related_log_id INTEGER NULL,
	type INTEGER NOT NULL,
	detail jsonb NOT NULL,
	FOREIGN KEY (user_id) REFERENCES "user"(id),
	FOREIGN KEY (related_log_id) REFERENCES user_log(id)
);

CREATE TABLE preform_test_a.foo (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	fk INTEGER NOT NULL --fk:preform_test_b.bar.id
);

CREATE TABLE preform_test_b.bar (
	id INTEGER PRIMARY KEY NOT NULL
);

`, wd, wd))
	if err != nil {
		panic(err)
	}
}
