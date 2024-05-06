package gormModel

import "time"

type TestA struct {
	Id     int64
	Name   string
	Int4   int32
	Int8   int64
	Float4 float32
	Float8 float64
	Bool   bool
	Text   string
	Time   time.Time
	TestBs []TestB `gorm:"foreignKey:AId"`
}

func (t TestA) TableName() string {
	return "preform_benchmark.test_a"
}

type TestB struct {
	Id     int64
	AId    int64
	Name   string
	Int4   int32
	Int8   int64
	Float4 float32
	Float8 float64
	Bool   bool
	Text   string
	Time   time.Time
	TestCs []TestC `gorm:"foreignKey:BId"`
}

func (t TestB) TableName() string {
	return "preform_benchmark.test_b"
}

type TestC struct {
	Id     int64
	BId    int64
	Name   string
	Int4   int32
	Int8   int64
	Float4 float32
	Float8 float64
	Bool   bool
	Text   string
	Time   time.Time
}

func (t TestC) TableName() string {
	return "preform_benchmark.test_c"
}
