package preformTypes

import (
	"database/sql/driver"
	"time"
)

type SqliteTime time.Time

func (t *SqliteTime) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch src.(type) {
	case time.Time:
		*t = SqliteTime(src.(time.Time))
	case []byte:
		tt, err := time.Parse("2006-01-02 15:04:05-07:00", string(src.([]byte)))
		if err != nil {
			return err
		}
		*t = SqliteTime(tt)
	case string:
		tt, err := time.Parse("2006-01-02 15:04:05-07:00", src.(string))
		if err != nil {
			return err
		}
		*t = SqliteTime(tt)
	case int64:
		*t = SqliteTime(time.Unix(src.(int64), 0))
	case float64:
		*t = SqliteTime(time.Unix(int64(src.(float64)), 0))

	}
	return nil
}

func (t SqliteTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}
