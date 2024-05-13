package preformShare

import (
	"context"
	"database/sql"
	"github.com/go-preform/squirrel"
	"github.com/jmoiron/sqlx"
	"reflect"
)

type IFactoryBuilder interface {
	CodeName() string
	FullCodeName() string
	SchemaName() string
	TableName() string
	FactoryType() reflect.Type
	SetAlias(alias string) IFactoryBuilder
	Alias() string
	ColSet() map[string]IColDef
	Cols() []IColDef
	PK() []IColDef
	Clone() IFactoryBuilder
	AddCode(def, model, setting, extraFuncs []string, setRelationCodes [][]string)
}

type IFactoryBuilderWithSetup interface {
	Setup() (skipAutoSetter bool)
}

type IQueryBuilder interface {
}

type IColDef interface {
	SrcName() string
	CodeName() string
	Alias() string
	SetAliasI(string) IColDef
	OColName() string
	GetType() reflect.Type
	Factory() IFactoryBuilder
	GenerateCode(schemaName string, fromQuery bool) (importPath []string, defColCode []string, modelColCode []string, settingCode []string, extraFunc string)
	ColDef() IColDef

	RelatedFks() []IColDef
	NewValue() any
}

type ICondForBuilder interface {
	ToCode() (code string)
	ToCondCode() (code string)
	IAnd(conds ...ICondForBuilder) ICondForBuilder
	UseAlias(...IColDef) ICondForBuilder
}

type IField interface {
	Name() string
	NewValue() any
	ParentModel() any
}

type ICol interface {
	IField
	DbName() string
	GetCode() string
	GetCodeWithAlias() string
	SetValue(any, any)
	GetPos() int
}
type Aggregator string
type SqlDialectLastInsertIdMethod uint32

type IDialect interface {
	QuoteIdentifier(string) string
	GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*Scheme
	Aggregate(fn Aggregator, body any, params ...any) squirrel.Sqlizer
	LastInsertIdMethod() (method SqlDialectLastInsertIdMethod, suffix func(col string) squirrel.Sqlizer)
	UpdateSqlizer(UpdateBuilder) (string, []any, error) //for clickhouse
	DeleteSqlizer(DeleteBuilder) (string, []any, error) //for clickhouse
	ValueParsers() map[reflect.Type]any
	DefaultValueExpr() squirrel.Sqlizer
	ParseCustomTypeScan(src any) (dst []string, err error)
	ParseCustomTypeValue(name string, src ...any) (dst string, err error)
	CaseStmtToSql(builder squirrel.CaseBuilder, col ICol) (string, []any, error)
	//condition
	Eq(col ICol, v any) squirrel.Sqlizer
	NotEq(col ICol, v any) squirrel.Sqlizer
	Like(col ICol, v any) squirrel.Sqlizer
	Gt(col ICol, v any) squirrel.Sqlizer
	GtOrEq(col ICol, v any) squirrel.Sqlizer
	Lt(col ICol, v any) squirrel.Sqlizer
	LtOrEq(col ICol, v any) squirrel.Sqlizer
	Between(col ICol, v1, v2 any) squirrel.Sqlizer

	ArrayEq(arrCol any, value any) (query string, args []any, err error)
	ArrayAny(arrCol any, value any) (query string, args []any, err error)
	ArrayHasAny(arrColA any, arrColB any) (query string, args []any, err error)
	ArrayConcat(arrColA any, arrColB any) (query string, args []any, err error)
	ArrayContains(arrColA any, arrColB any) (query string, args []any, err error)
	ArrayContainsBy(arrColA any, arrColB any) (query string, args []any, err error)
}

type ISqlizerWithDialect interface {
	WithDialect(iDialect IDialect) squirrel.Sqlizer
	ToSql(dialect IDialect) (string, []interface{}, error)
}

type IArrayTypes interface {
	IterAny() []any
}

type IQueryFactory interface {
	TableNames() []string
}

type DbQueryRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

type QueryRunner interface {
	DbQueryRunner
	InsertAndReturnAutoId(ctx context.Context, lastIdMethod SqlDialectLastInsertIdMethod, query string, args ...any) (int64, error)
	BaseRunner() DbQueryRunner
	RelatedFactory([]IQueryFactory) QueryRunner
}

type ITestQueryRunner interface {
	DbQueryRunner
	IsTester()
}

type ITestSqlConnectorDriver interface {
	TestDriverName() string
}

type IHasTypeForExport interface {
	TypeForExport() any
}
