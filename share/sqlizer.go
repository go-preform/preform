package preformShare

type SqlizerWithDialectWrapper struct {
	Sqlizer func() (string, []interface{}, error)
}

func (s SqlizerWithDialectWrapper) ToSql() (string, []interface{}, error) {
	return s.Sqlizer()
}
