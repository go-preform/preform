package main

//go test -v -bench . -benchmem -benchtime 2s

import (
	"database/sql"
	"fmt"
	"github.com/go-preform/preform"
	"github.com/jmoiron/sqlx"
	"testing"
)

/*
CREATE TABLE public.test_foo (
	id int8 NOT NULL,
	"name" varchar NOT NULL,
	"int4" int4 NOT NULL,
	"int8" int8 NOT NULL,
	"float4" float4 NOT NULL,
	"float8" float8 NOT NULL,
	"bool" bool NOT NULL,
	"text" text NOT NULL,
	CONSTRAINT test_foo_pk PRIMARY KEY (id)
);
*/

type Foo struct {
	Id     int
	Name   string
	Int4   int32
	Int8   int64
	Float4 float32
	Float8 float64
	Bool   bool
	Text   string
}

func (f *Foo) Fields() []interface{} {
	return []interface{}{&f.Id, &f.Name, &f.Int4, &f.Int8, &f.Float4, &f.Float8, &f.Bool, &f.Text}
}

var db *sqlx.DB

func init() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", 5432, "postgres", "123456", "postgres")

	//pgx.Connect(context.Background(), psqlconn)

	// open database
	conn, err := sql.Open("pgx", psqlconn)

	if err != nil {
		panic(err)
	}

	db = sqlx.NewDb(conn, "pgx")

	ts := &TestSchema{}

	ts.Foo = preform.InitFactory[*FooFactory, FooBody](func(d *TestSchema) {
		d.Foo.SetTableName("test_foo")
	})()
	ts.Init("public", ts, conn)

	preform.PrepareFactories(ts)

	fooFactory = ts.Foo

	//for i := 0; i < 15000; i++ {
	//	_, err := db.Exec("INSERT INTO test_foo(id, \"name\", \"int4\", \"float4\", \"bool\", \"int8\", \"float8\", \"text\")VALUES($7, $1, $2, $3, false, $4, $5, $6);",
	//		fmt.Sprintf("name_%d", i),
	//		i,
	//		float32(i)*1.01,
	//		i,
	//		float64(i)*1.01,
	//		fmt.Sprintf("text_%d", i),
	//		i+1,
	//	)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}

}

//func BenchmarkSqlxStructScan(b *testing.B) {
//	q, _ := db.Preparex("select id,name,int4,int8,float4,float8,bool,text from test_foo where id != $1 or \"name\" like $2;")
//	for i := 0; i < b.N; i++ {
//		rows, _ := q.Queryx(9999, "%name_%")
//
//		var (
//			foo  Foo
//			foos []Foo
//			err  error
//		)
//
//		for rows.Next() {
//			err = rows.StructScan(&foo)
//			if err != nil {
//				fmt.Println(err)
//			}
//			foos = append(foos, foo)
//		}
//		if foos[0].Id != 1 || foos[1000].Id != 1001 {
//			b.Fail()
//		}
//	}
//}
//
//func BenchmarkSqlxMapScan(b *testing.B) {
//	q, _ := db.Preparex("select id,name,int4,int8,float4,float8,bool,text from test_foo where id != $1 or \"name\" like $2;")
//	for i := 0; i < b.N; i++ {
//		rows, _ := q.Queryx(9999, "%name_%")
//
//		var (
//			m   map[string]interface{}
//			ms  []map[string]interface{}
//			err error
//		)
//
//		for rows.Next() {
//			m = make(map[string]interface{})
//			err = rows.MapScan(m)
//			if err != nil {
//				fmt.Println(err)
//			}
//			ms = append(ms, m)
//		}
//		if ms[0]["id"] != int64(1) || ms[1000]["id"] != int64(1001) {
//			b.Fail()
//		}
//	}
//}

//
//func BenchmarkArrScan(b *testing.B) {
//	q, _ := db.Prepare("select id,name,int4,int8,float4,float8,bool,text from test_foo")
//	for i := 0; i < b.N; i++ {
//		rows, _ := q.Query()
//
//		var (
//			foo  Foo
//			foos []Foo
//			err  error
//			arr  = make([]interface{}, 8)
//		)
//
//		arr[0], arr[1], arr[2], arr[3], arr[4], arr[5], arr[6], arr[7] = &foo.Id, &foo.DbName, &foo.Int4, &foo.Int8, &foo.Float4, &foo.Float8, &foo.Bool, &foo.Text
//
//		for rows.Next() {
//			err = rows.Scan(arr...)
//			if err != nil {
//				fmt.Println(err)
//			}
//			foos = append(foos, foo)
//		}
//		if foos[0].Id != 1 || foos[1000].Id != 1001 {
//			b.Fail()
//		}
//	}
//}
//
//func BenchmarkArrFieldsScan(b *testing.B) {
//	q, _ := db.Prepare("select id,name,int4,int8,float4,float8,bool,text from test_foo")
//	for i := 0; i < b.N; i++ {
//		rows, _ := q.Query()
//
//		var (
//			foo  Foo
//			foos []Foo
//			err  error
//			arr  []interface{}
//		)
//		arr = foo.Fields()
//
//		for rows.Next() {
//			err = rows.Scan(arr...)
//			if err != nil {
//				fmt.Println(err)
//			}
//			foos = append(foos, foo)
//		}
//		if foos[0].Id != 1 || foos[1000].Id != 1001 {
//			b.Fail()
//		}
//	}
//}

var fooFactory *FooFactory

type TestSchema struct {
	preform.Schema[*TestSchema, TestSchema]
	Foo *FooFactory
}

// Clone
func (t *TestSchema) Clone(name string, db ...*sql.DB) preform.ISchema {
	return &TestSchema{}
}

// GetFactories
func (t *TestSchema) Factories() []preform.IFactory {
	return []preform.IFactory{t.Foo}
}

type FooFactory struct {
	preform.Factory[*FooFactory, FooBody]
	Id     *preform.PrimaryKey[int]
	Name   *preform.Column[string]
	Int4   *preform.Column[int32]   `db:"int4"`
	Int8   *preform.Column[int64]   `db:"int8"`
	Float4 *preform.Column[float32] `db:"float4"`
	Float8 *preform.Column[float64] `db:"float8"`
	Bool   *preform.Column[bool]
	Text   *preform.Column[string]
}

type FooBody struct {
	preform.Body[FooBody, *FooFactory]
	Id     int
	Name   string
	Int4   int32   `db:"int4"`
	Int8   int64   `db:"int8"`
	Float4 float32 `db:"float4"`
	Float8 float64 `db:"float8"`
	Bool   bool
	Text   string
}

func (f *FooBody) FieldValues() []interface{} {
	return []interface{}{&f.Id, &f.Name, &f.Int4, &f.Int8, &f.Float4, &f.Float8, &f.Bool, &f.Text}
}

func (f *FooBody) FieldValuePtrs() []interface{} {
	return []interface{}{&f.Id, &f.Name, &f.Int4, &f.Int8, &f.Float4, &f.Float8, &f.Bool, &f.Text}
}

func (f FooBody) Factory() *FooFactory {
	return nil
}

func BenchmarkPreheat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		foo := Foo{}
		foo.Id = 1
		foo.Name = "name"
		foo.Int4 = 1
		foo.Int8 = 1
		foo.Float4 = 1
		foo.Float8 = 1
		foo.Bool = true
		foo.Text = "text"
	}
}

func BenchmarkPreformScanFast(b *testing.B) {
	q, err := fooFactory.Select(
		fooFactory.Text,
		fooFactory.Bool,
		fooFactory.Float8,
		fooFactory.Float4,
		fooFactory.Int8,
		fooFactory.Int4,
		fooFactory.Name,
		fooFactory.Id,
	).Limit(1001).Prepare()
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < b.N; i++ {
		foos, err := q.GetAllFast()
		if err != nil {
			fmt.Println(err)
		}
		if foos[0].Id != 1 || foos[1000].Id != 1001 {
			b.Fail()
		}
	}
}

func BenchmarkPreformScan(b *testing.B) {
	q, err := fooFactory.Select(
		fooFactory.Text,
		fooFactory.Bool,
		fooFactory.Float8,
		fooFactory.Float4,
		fooFactory.Int8,
		fooFactory.Int4,
		fooFactory.Name,
		fooFactory.Id,
	).Limit(1001).Prepare()
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < b.N; i++ {
		foos, err := q.GetAll()
		if err != nil {
			fmt.Println(err)
		}
		if foos[0].Id != 1 || foos[1000].Id != 1001 {
			b.Fail()
		}
	}
}

func BenchmarkHardCodeScan(b *testing.B) {
	q, _ := db.Prepare("select id,name,int4,int8,float4,float8,bool,text from test_foo limit 1001;")
	for i := 0; i < b.N; i++ {
		rows, _ := q.Query()

		var (
			foo  Foo
			foos []Foo
			ptrs = []interface{}{&foo.Id, &foo.Name, &foo.Int4, &foo.Int8, &foo.Float4, &foo.Float8, &foo.Bool, &foo.Text}
			err  error
		)

		for rows.Next() {
			err = rows.Scan(ptrs...)
			if err != nil {
				fmt.Println(err)
			}
			foos = append(foos, foo)
		}
		if foos[0].Id != 1 || foos[1000].Id != 1001 {
			b.Fail()
		}
	}
}
