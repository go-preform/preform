package preformShare

import (
	"database/sql"
	"github.com/go-preform/squirrel"
)

type (
	SelectBuilder = squirrel.SelectBuilderFast
	DeleteBuilder = squirrel.DeleteBuilderFast
	UpdateBuilder = squirrel.UpdateBuilderFast
	//InsertBuilder = squirrel.InsertBuilderFast
	Scheme struct {
		Tables      []*Table
		Name        string
		Imports     map[string]struct{}
		Enums       map[string][]string
		CustomTypes map[string]*CustomType
	}

	CustomTypeAttr struct {
		Name      string
		Type      string
		NotNull   bool
		IsScanner bool
	}
	CustomType struct {
		Name    string
		Attr    []*CustomTypeAttr
		Imports map[string]struct{}
	}
	Table struct {
		Name          string
		Columns       []*Column
		Comment       string
		Scheme        *Scheme            `json:"-"`
		ColumnByName  map[string]*Column `json:"-"`
		Imports       map[string]struct{}
		Inheritors    [][3]string
		IsView        bool
		IsMiddleTable bool
		ForeignKeys   map[string]*ForeignKey
		Sql           string
	}
	Column struct {
		Name         string
		Type         string
		GoType       string
		Nullable     bool
		IsPrimaryKey bool
		PkPos        int64
		IsAutoKey    bool
		Comment      string
		ForeignKeys  []*ForeignKey
		Table        *Table `json:"-"`
		DefaultValue sql.NullString
		IColDef      IColDef
		IsScanner    bool
	}
	ForeignKey struct {
		Name         string
		LocalKeys    []*Column
		ForeignKeys  []*Column
		RelationName string
		ReverseName  string
		AssociatedFk *ForeignKey
	}
)
