package main

import (
	"context"
	"database/sql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/go-preform/preform/benchmark/config"
	"github.com/go-preform/preform/benchmark/ent"
	gormModel "github.com/go-preform/preform/benchmark/gorm"
	"github.com/go-preform/preform/benchmark/preform/preformModel"
	"github.com/jmoiron/sqlx"
	gormPg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

var (
	pgConn    *sql.DB
	gormConn  *gorm.DB
	entClient *ent.Client
	sqlxConn  *sqlx.DB
)

func init() {
	var err error
	pgConn, err = sql.Open("pgx", config.PgConnStr)
	if err != nil {
		panic(err)
	}
	_, err = pgConn.Exec(`
	DROP SCHEMA IF EXISTS preform_benchmark CASCADE;
	CREATE SCHEMA preform_benchmark;
	CREATE TABLE preform_benchmark.test_a (
		id int8 NOT NULL,
		"name" varchar NOT NULL,
		"int4" int4 NOT NULL,
		"int8" int8 NOT NULL,
		"float4" float4 NOT NULL,
		"float8" float8 NOT NULL,
		"bool" bool NOT NULL,
		"text" text NOT NULL,
	"time" timestamptz NOT NULL,
		CONSTRAINT test_a_pk PRIMARY KEY (id)
	);
	
	CREATE TABLE preform_benchmark.test_b (
		id int8 NOT NULL,
		a_id int8 not null,
		"name" varchar NOT NULL,
		"int4" int4 NOT NULL,
		"int8" int8 NOT NULL,
		"float4" float4 NOT NULL,
		"float8" float8 NOT NULL,
		"bool" bool NOT NULL,
		"text" text NOT NULL,
	"time" timestamptz NOT NULL,
		CONSTRAINT test_b_pk PRIMARY KEY (id),
		CONSTRAINT test_b_pk_fk FOREIGN KEY (a_id) REFERENCES preform_benchmark."test_a"(id)
	);
	
	CREATE TABLE preform_benchmark.test_c (
		id int8 NOT NULL,
		b_id int8 not null,
		"name" varchar NOT NULL,
		"int4" int4 NOT NULL,
		"int8" int8 NOT NULL,
		"float4" float4 NOT NULL,
		"float8" float8 NOT NULL,
		"bool" bool NOT NULL,
		"text" text NOT NULL,
	"time" timestamptz NOT NULL,
		CONSTRAINT test_c_pk PRIMARY KEY (id),
		CONSTRAINT test_c_pk_fk FOREIGN KEY (b_id) REFERENCES preform_benchmark."test_b"(id)
	);`)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 1000; i++ {
		_, err := pgConn.Exec("INSERT INTO preform_benchmark.test_a(id, \"name\", \"int4\", \"float4\", \"bool\", \"int8\", \"float8\", \"text\", \"time\")VALUES($7, $1, $2, $3, false, $4, $5, $6,$8);",
			fmt.Sprintf("name_%d", i),
			i,
			float32(i)*1.01,
			i,
			float64(i)*1.01,
			fmt.Sprintf("text_%d", i),
			i+1,
			time.Now(),
		)
		if err != nil {
			fmt.Println(err)
		}
		_, err = pgConn.Exec("INSERT INTO preform_benchmark.test_b(id, a_id, \"name\", \"int4\", \"float4\", \"bool\", \"int8\", \"float8\", \"text\", \"time\")VALUES($7, $8, $1, $2, $3, false, $4, $5, $6, $9);",
			fmt.Sprintf("name_%d", i),
			i,
			float32(i)*1.01,
			i,
			float64(i)*1.01,
			fmt.Sprintf("text_%d", i),
			i+1,
			i%50+1,
			time.Now(),
		)
		if err != nil {
			fmt.Println(err)
		}
		_, err = pgConn.Exec("INSERT INTO preform_benchmark.test_c(id, b_id, \"name\", \"int4\", \"float4\", \"bool\", \"int8\", \"float8\", \"text\", \"time\")VALUES($7, $8, $1, $2, $3, false, $4, $5, $6, $9);",
			fmt.Sprintf("name_%d", i),
			i,
			float32(i)*1.01,
			i,
			float64(i)*1.01,
			fmt.Sprintf("text_%d", i),
			i+1,
			i%50+1,
			time.Now(),
		)
		if err != nil {
			fmt.Println(err)
		}
	}
	preformModel.Init(pgConn)
	gormConn, err = gorm.Open(gormPg.Open(config.PgConnStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	drv := entsql.OpenDB(dialect.Postgres, pgConn)
	entClient = ent.NewClient(ent.Driver(drv))
	_, err = pgConn.Exec("SET SEARCH_PATH = preform_benchmark")
	if err != nil {
		panic(err)
	}
	sqlxConn = sqlx.NewDb(pgConn, "pgx")
}

func BenchmarkPreformSelectAll(b *testing.B) {
	var (
		testAs []preformModel.TestABody
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = preformModel.PreformBenchmark.TestA.Select().Limit(1000).GetAll()
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 1000 {
			b.Error("len(testAs) != 1000")
		}
	}
}

func BenchmarkPreformSelectAllFast(b *testing.B) {
	var (
		testAs []preformModel.TestABody
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = preformModel.PreformBenchmark.TestA.Select().Limit(1000).GetAllFast()
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 1000 {
			b.Error("len(testAs) != 1000")
		}
	}
}

func BenchmarkPreformSelectEager(b *testing.B) {
	var (
		testAs []preformModel.TestABody
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = preformModel.PreformBenchmark.TestA.Select().Eager(preformModel.PreformBenchmark.TestA.TestBS.Eager(preformModel.PreformBenchmark.TestB.TestCS)).Limit(100).GetAll()
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 100 {
			b.Error("len(testAs) != 1000")
		}
		if len(testAs[0].TestBS) != 20 {
			b.Error("len(testAs[0].TestBs) != 20")
		}
	}
}

func BenchmarkPreformSelectEagerFast(b *testing.B) {
	var (
		testAs []preformModel.TestABody
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = preformModel.PreformBenchmark.TestA.Select().Eager(preformModel.PreformBenchmark.TestA.TestBS.Eager(preformModel.PreformBenchmark.TestB.TestCS)).Limit(100).GetAllFast()
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 100 {
			b.Error("len(testAs) != 1000")
		}
		if len(testAs[0].TestBS) != 20 {
			b.Error("len(testAs[0].TestBs) != 20")
		}
	}
}

func BenchmarkGormSelectAll(b *testing.B) {
	var (
		testAs []gormModel.TestA
		err    error
	)
	for i := 0; i < b.N; i++ {
		err = gormConn.Limit(1000).Find(&testAs).Error
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 1000 {
			b.Error("len(testAs) != 1000")
		}
	}
}

func BenchmarkGormSelectEager(b *testing.B) {
	var (
		testAs []gormModel.TestA
		err    error
	)
	for i := 0; i < b.N; i++ {
		err = gormConn.Preload("TestBs.TestCs").Limit(100).Find(&testAs).Error
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 100 {
			b.Error("len(testAs) != 100")
		}
		if len(testAs[0].TestBs) != 20 {
			b.Error("len(testAs[0].TestBs) != 20")
		}
	}
}

func BenchmarkEntSelectAll(b *testing.B) {
	var (
		testAs []*ent.TestA
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = entClient.TestA.Query().Limit(1000).All(context.Background())
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 1000 {
			b.Error("len(testAs) != 1000")
		}
	}
}

func BenchmarkEntSelectEager(b *testing.B) {
	var (
		testAs []*ent.TestA
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs, err = entClient.TestA.Query().WithTestBs(func(query *ent.TestBQuery) {
			query.WithTestCs()
		}).Limit(100).All(context.Background())
		if err != nil {
			b.Error(err)
		}
		if len(testAs) != 100 {
			b.Error("len(testAs) != 100")
		}
		if len(testAs[0].Edges.TestBs) != 20 {
			b.Error("len(testAs[0].TestBs) != 20")
		}
	}
}

func BenchmarkSqlxStructScan(b *testing.B) {
	var (
		testAs []preformModel.TestABody
		testA  preformModel.TestABody
		rows   *sqlx.Rows
		err    error
	)
	for i := 0; i < b.N; i++ {
		testAs = nil
		rows, err = sqlxConn.Queryx("SELECT id, \"name\", \"int4\", \"int8\", \"float4\", \"float8\", \"bool\", \"text\", \"time\" FROM preform_benchmark.test_a limit 1000;")
		if err != nil {
			b.Error(err)
		}
		for rows.Next() {
			err = rows.StructScan(&testA)
			if err != nil {
				b.Error(err)
			}
			testAs = append(testAs, testA)
		}
		if len(testAs) != 1000 {
			fmt.Println(len(testAs))
			b.Error("len(testAs) != 1000")
		}
	}
}

func BenchmarkSqlRawScan(b *testing.B) {

	for i := 0; i < b.N; i++ {
		var (
			testAs = make([]preformModel.TestABody, 0, 1000)
			testA  preformModel.TestABody
			rows   *sql.Rows
			err    error
			ptrs   = []any{&testA.Id, &testA.Name, &testA.Int4, &testA.Int8, &testA.Float4, &testA.Float8, &testA.Bool, &testA.Text, &testA.Time}
		)
		rows, err = sqlxConn.Query("SELECT id, \"name\", \"int4\", \"int8\", \"float4\", \"float8\", \"bool\", \"text\", \"time\" FROM preform_benchmark.test_a limit 1000;")
		if err != nil {
			b.Error(err)
		}
		for rows.Next() {
			err = rows.Scan(ptrs...)
			if err != nil {
				b.Error(err)
			}
			testAs = append(testAs, testA)
		}
		if len(testAs) != 1000 {
			fmt.Println(len(testAs))
			b.Error("len(testAs) != 1000")
		}
	}
}
