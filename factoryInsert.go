package preform

import (
	"context"
	"encoding/json"
	"errors"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
)

var (
	noFixCondErr = errors.New("insert with fix condition mirror factory is not supported")
)

type EditConfig struct {
	Tx               *Tx
	Cascading        bool
	Ctx              context.Context
	NoAutoPrimaryKey bool
}

// Insert body bodies map maps with cascading and insert id
func (f *Factory[FPtr, B]) Insert(body any, cfg ...EditConfig) (err error) {
	switch body.(type) {
	case B:
		var b = body.(B)
		return f.InsertOne(&b, cfg...)
	case *B:
		return f.InsertOne(body.(*B), cfg...)
	case []*B:
		for _, b := range body.([]*B) {
			if err := f.InsertOne(b, cfg...); err != nil {
				return err
			}
		}
	case []B:
		for _, b := range body.([]B) {
			if err := f.InsertOne(&b, cfg...); err != nil {
				return err
			}
		}
	case []any:
		for _, b := range body.([]any) {
			if err = f.Insert(b, cfg...); err != nil {
				return err
			}
		}
	case map[string]any:
		return f.InsertMap(body.(map[string]any), cfg...)
	case []map[string]any:
		for _, b := range body.([]map[string]any) {
			if err = f.InsertMap(b, cfg...); err != nil {
				return err
			}
		}
	}
	return nil
}

// for []*B, []B insert not require insert id and cascading
func (f Factory[FPtr, B]) InsertBatch(bodies any, cfgs ...EditConfig) error {
	if f.fixCond != nil {
		return errors.New("insert with fix condition mirror factory is not supported")
	}
	var (
		db             = f.Db()
		exec        DB = db
		query          = db.sqStmtBuilder.InsertFast(f.tableNameWithParent())
		chunkQuery     = query
		cfg         EditConfig
		ctx         = db.ctx
		autoPk      = f.autoPk
		loopInsert  func(body iModelBodyReadOnly, i int) bool
		returnErr   error
		defaultExpr = db.dialect.DefaultValueExpr()
	)
	if defaultExpr == nil {
		return f.Insert(bodies, cfgs...) //no default value driver can't go batch
	}
	if len(cfgs) != 0 {
		cfg = cfgs[0]
	}
	if cfg.Tx != nil {
		exec = cfg.Tx
	}
	if cfg.Ctx != nil {
		ctx = cfg.Ctx
	}
	if cfg.Cascading {
		return errors.New("cascading is not supported for batch insert")
	}
	if cfg.NoAutoPrimaryKey {
		autoPk = nil
	}
	for _, col := range f.columns {
		if autoPk == col {
			continue
		}
		query = query.Columns(db.dialect.QuoteIdentifier(col.DbName()))
	}

	execSql := func() error {
		q, args, _ := chunkQuery.ToSql()
		if len(args) == 0 {
			return nil
		}
		_, err := exec.RelatedFactory([]preformShare.IQueryFactory{f}).ExecContext(ctx, q, args...)
		if err != nil {
			return err
		}
		chunkQuery = query
		return nil
	}

	if autoPk == nil {
		loopInsert = func(body iModelBodyReadOnly, i int) bool {
			insertBodyValues := body.FieldValueImmutablePtrs()
			for i, v := range insertBodyValues {
				insertBodyValues[i] = f.columns[i].unwrapPtrForInsert(v)
				if insertBodyValues[i] == preformShare.DEFAULT_VALUE {
					insertBodyValues[i] = db.dialect.DefaultValueExpr()
				}
			}
			chunkQuery = chunkQuery.Values(insertBodyValues...)
			if i != 0 && i%INSERT_CHUNK_SIZE == 0 {
				err := execSql()
				if err != nil {
					returnErr = err
					return false
				}
				chunkQuery = query
			}
			return true
		}

	} else {
		loopInsert = func(body iModelBodyReadOnly, i int) bool {
			var (
				bodyValues       = body.FieldValueImmutablePtrs()
				insertBodyValues = make([]any, len(bodyValues)-1)
				pk               = autoPk.GetPos()
			)
			for i, v := range bodyValues {
				if i < pk {
					insertBodyValues[i] = f.columns[i].unwrapPtrForInsert(v)
					if insertBodyValues[i] == preformShare.DEFAULT_VALUE {
						insertBodyValues[i] = db.dialect.DefaultValueExpr()
					}
				} else if i != pk {
					insertBodyValues[i-1] = f.columns[i].unwrapPtrForInsert(v)
					if insertBodyValues[i-1] == preformShare.DEFAULT_VALUE {
						insertBodyValues[i-1] = db.dialect.DefaultValueExpr()
					}
				}
			}
			chunkQuery = chunkQuery.Values(insertBodyValues...)
			if i != 0 && i%INSERT_CHUNK_SIZE == 0 {
				err := execSql()
				if err != nil {
					returnErr = err
					return false
				}
				chunkQuery = query
			}
			return true
		}
	}

	if returnErr != nil {
		return returnErr
	}

	chunkQuery = query
	switch bodies.(type) {
	case []B:
		loopSlice[B, iModelBodyReadOnly](bodies.([]B), loopInsert)
	case []*B:
		loopSlice[*B, iModelBodyReadOnly](bodies.([]*B), loopInsert)
	default:
		return errors.New("bodies is not []B or []*B")
	}
	returnErr = execSql()
	if returnErr != nil {
		return returnErr
	}
	return returnErr
}

func (f Factory[FPtr, B]) InsertOne(body *B, cfgs ...EditConfig) error {
	if f.fixCond != nil {
		return noFixCondErr
	}
	var (
		db              = f.Db()
		exec         DB = db
		query           = db.sqStmtBuilder.InsertFast(f.tableNameWithParent())
		bodyValues      = any(body).(iModelBody).FieldValuePtrs()
		cfg          EditConfig
		ctx          = f.Db().ctx
		autoPk       = f.autoPk
		lastIdMethod preformShare.SqlDialectLastInsertIdMethod
		lastIdSuffix func(col string) squirrel.Sqlizer
	)
	lastIdMethod, lastIdSuffix = db.dialect.LastInsertIdMethod()
	if len(cfgs) != 0 {
		cfg = cfgs[0]
	}
	if cfg.Tx != nil {
		exec = cfg.Tx
	}
	if cfg.Ctx != nil {
		ctx = cfg.Ctx
	}
	if cfg.NoAutoPrimaryKey {
		autoPk = nil
	}
	var i int
	for _, col := range f.columns {
		bodyValues[i] = col.unwrapPtrForInsert(bodyValues[i])
		if bodyValues[i] == preformShare.DEFAULT_VALUE {
			if autoPk == col {
				bodyValues = append(bodyValues[:autoPk.GetPos()], bodyValues[autoPk.GetPos()+1:]...)
				if lastIdSuffix != nil {
					suffix, args, _ := lastIdSuffix(autoPk.DbName()).ToSql()
					query = query.Suffix(suffix, args...)
				}
				continue
			} else {
				bodyValues = append(bodyValues[:i], bodyValues[i+1:]...) //sqlite don't have DEFAULT
				continue
			}
		}
		query = query.Columns(db.dialect.QuoteIdentifier(col.DbName()))
		i++
	}
	query = query.Values(bodyValues[:i]...)
	q, args, err := query.ToSql()
	if err != nil {
		return err
	}
	if autoPk == nil {
		_, err = exec.RelatedFactory([]preformShare.IQueryFactory{f}).ExecContext(ctx, q, args...)
		if err != nil {
			return err
		}
	} else {
		var (
			lastId int64
		)
		lastId, err = exec.RelatedFactory([]preformShare.IQueryFactory{f}).InsertAndReturnAutoId(ctx, lastIdMethod, q, args...)
		if err != nil {
			return err
		}
		setInsertId(body, autoPk, lastId)
	}
	if cfg.Cascading {
		if len(f.relations) != 0 {
			var (
				bodyAsiModelRelatedBody = any(body).(iModelRelatedBody)
				relatedBodies           []any
				relatedBody             any
			)
			for _, rel := range f.relations {
				relatedBodies = rel.unwrapPtrBodyToTargetBodies(bodyAsiModelRelatedBody.RelatedValuePtrs()[rel.Index()])
				if relatedBodies != nil {
					for _, relatedBody = range relatedBodies {
						if relatedBody != nil {
							if rel.IsMiddleTable() {
								if err = rel.TargetFactory().Insert(relatedBody, cfgs...); err != nil {
									return err
								}
								if err = rel.setForeignKey(body, relatedBody); err != nil {
									return err
								}
							} else {
								_ = rel.setForeignKey(body, relatedBody)
								if err = rel.TargetFactory().Insert(relatedBody, cfgs...); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}
	return nil

}

func (f Factory[FPtr, B]) InsertMap(body map[string]any, cfgs ...EditConfig) error {
	if f.fixCond != nil {
		return noFixCondErr
	}
	jsonStr, _ := json.Marshal(body)
	var (
		b    B
		bPtr = &b
		err  error
	)
	if err = json.Unmarshal(jsonStr, bPtr); err != nil {
		return err
	}
	err = f.InsertOne(bPtr, cfgs...)
	if f.autoPk != nil {
		body[f.autoPk.DbName()] = f.autoPk.getValueFromBodyFlatten(any(bPtr).(iModelBody))
	}
	return err
}

func setInsertId[B any](body *B, autoPk ICol, lastId int64) {
	var (
		va = autoPk.NewValue()
		id any
	)
	switch va.(type) {
	case int64:
		id = lastId
	case int:
		id = int(lastId)
	case int32:
		id = int32(lastId)
	case uint64:
		id = uint64(lastId)
	case uint32:
		id = uint32(lastId)
	case uint:
		id = uint(lastId)
	case float64:
		id = float64(lastId)
	case float32:
		id = float32(lastId)
	}
	autoPk.SetValue(body, id)
}
