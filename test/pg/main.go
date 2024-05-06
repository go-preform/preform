package main

import (
	"database/sql"
	"fmt"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/test/pg/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"os/exec"
)

func main() {
	pgxConn, _ := sql.Open("pgx", config.PgConnStr)
	prepareSchema(pgxConn)
	preformBuilder.BuildModel(pgxConn, "mainModel", "mainModel", "preform_test_a", "preform_test_b")

	fmt.Println("go test----------------------------")
	d, _ := os.Getwd()
	p := fmt.Sprintf("%s/model_test", d)
	cmd := exec.Command("go", "test", p)
	cmd.Dir = d
	out, err := cmd.CombinedOutput()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println(string(out), err)
}

func prepareSchema(pgxConn *sql.DB) {
	_, err := pgxConn.Exec("DROP SCHEMA IF EXISTS preform_test_a CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = pgxConn.Exec("DROP SCHEMA IF EXISTS preform_test_b CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = pgxConn.Exec(`
CREATE SCHEMA preform_test_a;
CREATE SCHEMA preform_test_b;

CREATE TYPE preform_test_a.log_type AS ENUM (
    'Register',
    'Login'
);

CREATE TYPE preform_test_a.log_detail AS (
    user_agent text,
    session_id uuid,
    last_login timestamptz
);

CREATE TABLE preform_test_a."user" (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	manager_ids _int4 NOT NULL,
	created_by int4 NOT NULL,
	created_at timestamptz NOT NULL,
	logined_at timestamptz NULL,
	detail jsonb DEFAULT NULL,
	config jsonb NOT NULL DEFAULT '{}'::json,
	extra_config jsonb NOT NULL DEFAULT '{}'::json,
	CONSTRAINT manager_pk PRIMARY KEY (id),
	CONSTRAINT user_user_fk FOREIGN KEY (created_by) REFERENCES preform_test_a."user"(id)
);

CREATE TABLE preform_test_a.user_log (
	id uuid NOT NULL,
	user_id int4 NOT NULL,
	related_log_id uuid NULL,
	type preform_test_a.log_type NOT NULL,
	detail preform_test_a.log_detail NOT NULL,
	CONSTRAINT user_log_pk PRIMARY KEY (id),
	CONSTRAINT user_log_user_fk FOREIGN KEY (user_id) REFERENCES preform_test_a."user"(id),
	CONSTRAINT user_log_user_log_fk FOREIGN KEY (related_log_id) REFERENCES preform_test_a.user_log(id)
);

CREATE TABLE preform_test_a.foo (
	id serial4 NOT NULL,
	fk1 int4 NOT NULL,
	fk2 int4 NOT NULL,
	CONSTRAINT foo_pk PRIMARY KEY (id),
	CONSTRAINT foo_un UNIQUE (fk1, fk2)
);

CREATE TABLE preform_test_b.bar (
	id1 int4 NOT NULL,
	id2 int4 NOT NULL,
	CONSTRAINT target_pk PRIMARY KEY (id1, id2)
);

ALTER TABLE preform_test_b.bar ADD CONSTRAINT bar_fk FOREIGN KEY (id1,id2) REFERENCES preform_test_a.foo(fk1,fk2);

`)
	if err != nil {
		panic(err)
	}
}
