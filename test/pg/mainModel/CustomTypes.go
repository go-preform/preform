package mainModel

import (
	"github.com/go-preform/preform/types"
	"github.com/satori/go.uuid"
	"time"
	"database/sql/driver"
)
type PreformTestALogDetail struct{
	UserAgent string
	SessionId uuid.UUID
	LastLogin time.Time
}

func (ct *PreformTestALogDetail) Scan(src any) error {
	inputs, err := PreformTestA.DB.GetDialect().ParseCustomTypeScan(src)
	if err != nil {
		return err
	}
	
	err = preformTypes.GenericScan(&ct.UserAgent, inputs[0])
	if err != nil {
		return err
	}

	err = ct.SessionId.Scan(inputs[1])
	if err != nil {
		return err
	}

	err = preformTypes.GenericScan(&ct.LastLogin, inputs[2])
	if err != nil {
		return err
	}

	return nil
}

func (ct PreformTestALogDetail) Value() (driver.Value, error) {
	return PreformTestA.DB.GetDialect().ParseCustomTypeValue("log_detail", ct.UserAgent, ct.SessionId, ct.LastLogin)
}



