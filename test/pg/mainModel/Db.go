package mainModel
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

func Init(conn *sql.DB, queryRunnerForTest ... preformShare.QueryRunner) {
	schemas := []preform.ISchema{}
	schemas = append(schemas, initPreformTestA(conn, "", queryRunnerForTest...))
	schemas = append(schemas, initPreformTestB(conn, "", queryRunnerForTest...))
	preform.PrepareQueriesAndRelation(schemas...)
}

func CloneAll(preformTestAName string, preformTestBName string, db ... *sql.DB) (preformTestA *PreformTestASchema, preformTestB *PreformTestBSchema) {
	preformTestA = PreformTestA.clone(preformTestAName, db...).(*PreformTestASchema)
	preformTestB = PreformTestB.clone(preformTestBName, db...).(*PreformTestBSchema)
	preform.PrepareQueriesAndRelation(preformTestA, preformTestB)
	preformTestA.Inherit(PreformTestA)
	preformTestB.Inherit(PreformTestB)
	return
}

