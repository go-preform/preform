package preformModel
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

func Init(conn *sql.DB, queryRunnerForTest ... preformShare.QueryRunner) {
	schemas := []preform.ISchema{}
	schemas = append(schemas, initPreformBenchmark(conn, "", queryRunnerForTest...))
	preform.PrepareQueriesAndRelation(schemas...)
}

func CloneAll(preformBenchmarkName string, db ... *sql.DB) (preformBenchmark *PreformBenchmarkSchema) {
	preformBenchmark = PreformBenchmark.clone(preformBenchmarkName, db...).(*PreformBenchmarkSchema)
	preform.PrepareQueriesAndRelation(preformBenchmark)
	preformBenchmark.Inherit(PreformBenchmark)
	return
}

