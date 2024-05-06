package mainModel
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

func Init(conn *sql.DB, queryRunnerForTest ... preformShare.QueryRunner) {
	schemas := []preform.ISchema{}
	schemas = append(schemas, initPreformTestA(conn, "", queryRunnerForTest...))
	preform.PrepareQueriesAndRelation(schemas...)
}

func CloneAll(preformTestAName string, db ... *sql.DB) (preformTestA *PreformTestASchema) {
	preformTestA = PreformTestA.clone(preformTestAName, db...).(*PreformTestASchema)
	preform.PrepareQueriesAndRelation(preformTestA)
	preformTestA.Inherit(PreformTestA)
	return
}

