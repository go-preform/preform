package main

import (
	"database/sql"
	"fmt"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/test/mysql/config"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"os/exec"
)

func main() {
	myConn, err := sql.Open("mysql", config.MysqlConnStr)
	if err != nil {
		panic(err)
	}
	prepareSchema(myConn)
	preformBuilder.BuildModel(myConn, "mainModel", "mainModel", "preform_test_a", "preform_test_b")

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
	_, err := myConn.Exec("SET FOREIGN_KEY_CHECKS=0;")
	if err != nil {
		panic(err)
	}
	_, err = myConn.Exec("DROP DATABASE IF EXISTS preform_test_a")
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
	_, err = myConn.Exec("SET FOREIGN_KEY_CHECKS=1;")
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.user (
	id int NOT NULL AUTO_INCREMENT,
	` + "`name`" + ` varchar(255) NOT NULL,
	created_by int NOT NULL,
	created_at datetime NOT NULL,
	logined_at datetime NULL,
  	PRIMARY KEY (id),
  CONSTRAINT user_FK FOREIGN KEY (created_by) REFERENCES preform_test_a.user (id)
);`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.user_manager (
		user_id int NOT NULL,
		manager_id int NOT NULL,
		CONSTRAINT user_manager_PK PRIMARY KEY (user_id,manager_id),
	  CONSTRAINT staff_FK FOREIGN KEY (user_id) REFERENCES preform_test_a.user (id) ON DELETE CASCADE,
	  CONSTRAINT manager_FK FOREIGN KEY (manager_id) REFERENCES preform_test_a.user (id) ON DELETE CASCADE
	);`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.user_log (
	id bigint NOT NULL AUTO_INCREMENT,
	user_id int NOT NULL,
	related_log_id bigint NULL,
	type enum('Register','login') NOT NULL,
	PRIMARY KEY (id),
  KEY user_log_FK (user_id),
  CONSTRAINT user_log_FK FOREIGN KEY (user_id) REFERENCES preform_test_a.user (id),
  CONSTRAINT user_log_related_FK FOREIGN KEY (related_log_id) REFERENCES preform_test_a.user_log (id)
);`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_a.foo (
	id int NOT NULL AUTO_INCREMENT,
	fk1 int NOT NULL,
	fk2 int NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY foo_un (fk1, fk2) USING BTREE
);`)
	if err != nil {
		panic(err)
	}

	_, err = myConn.Exec(`CREATE TABLE preform_test_b.bar (
	id1 int4 NOT NULL,
	id2 int4 NOT NULL,
	PRIMARY KEY (id1, id2),
  CONSTRAINT bar_FK FOREIGN KEY (id1, id2) REFERENCES preform_test_a.foo (fk1, fk2)
);`)
	if err != nil {
		panic(err)
	}
}
