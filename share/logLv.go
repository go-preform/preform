package preformShare

import "strconv"

type LogLv uint32

type ILogLv interface {
	UnmarshalJSON(b []byte) error
	Lv() uint32
}

const (
	LogLv_Exec   LogLv = 1
	LogLv_Read   LogLv = 2
	LogLv_Health LogLv = 4
)

func (o LogLv) Lv() uint32 {
	return uint32(o)
}

func (o *LogLv) UnmarshalJSON(b []byte) error {
	i, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*o = LogLv(i)
	return nil
}
