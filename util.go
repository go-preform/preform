package preform

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"strings"
)

func loopSlice[Src, Target any](slice []Src, fn func(Target, int) bool) {
	for i, v := range slice {
		if !fn(any(v).(Target), i) {
			return
		}
	}
}

func sqlizerToString(s squirrel.Sqlizer) string {
	str, args, _ := s.ToSql()
	if args != nil {
		str, args, _ = preformShare.NestSql(str, args)
		for _, arg := range args {
			if _, ok := arg.(string); ok {
				str = strings.Replace(str, "?", fmt.Sprintf(`"%s"`, arg.(string)), 1)
			} else {
				str = strings.Replace(str, "?", fmt.Sprintf(`%v`, arg), 1)
			}
		}
	}
	return str
}

func iterAny[T any](arr []T) []any {
	var res []any
	for _, v := range arr {
		res = append(res, v)
	}
	return res
}
