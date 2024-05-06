package preformBuilder

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type ConditionForBuilder[T any] struct {
	column     preformShare.IColDef
	conditions [][]any
	useAlias   []preformShare.IColDef
}

func NewConditionForBuilder[T any](column preformShare.IColDef) ConditionForBuilder[T] {
	return ConditionForBuilder[T]{column: column}
}

func (c ConditionForBuilder[T]) UseAlias(cols ...preformShare.IColDef) preformShare.ICondForBuilder {
	c.useAlias = cols
	return c
}

func (c ConditionForBuilder[T]) ToCondCode() string {
	var (
		condCodes []string
	)
	for _, cc := range c.conditions {
		argsCodes := []string{}
	colLoop:
		for _, a := range cc[1:] {
			if col, ok := a.(preformShare.IColDef); ok {
				if len(c.useAlias) != 0 {
					if col.Alias() == col.OColName() {
						for _, ua := range c.useAlias {
							if ua.ColDef() == col.ColDef() {
								argsCodes = append(argsCodes, fmt.Sprintf("s.%s.%s", col.Factory().CodeName(), strcase.ToCamel(col.Factory().CodeName()+" "+col.SrcName())))
								break colLoop
							}
						}
						argsCodes = append(argsCodes, fmt.Sprintf("s.%s.%s.%s", col.Factory().CodeName(), col.Factory().Alias(), col.SrcName()))
					} else {
						argsCodes = append(argsCodes, fmt.Sprintf("s.%s.%s", col.Factory().CodeName(), col.Alias()))
					}
				} else {
					argsCodes = append(argsCodes, fmt.Sprintf("s.%s.%s.%s", col.Factory().CodeName(), col.Factory().Alias(), col.SrcName()))
				}
			} else {
				t := reflect.TypeOf(a)
				switch t.Kind() {
				case reflect.String:
					argsCodes = append(argsCodes, fmt.Sprintf("\"%s\"", a))
				case reflect.Array, reflect.Slice:
					if cc[0] == "And" || cc[0] == "Or" {

						var (
							conds = a.([]preformShare.ICondForBuilder)
						)
						for _, cond := range conds {
							argsCodes = append(argsCodes, cond.ToCondCode())
						}
					} else {
						var (
							v  = reflect.ValueOf(a)
							vs = make([]any, v.Len())
						)
						for i := 0; i < v.Len(); i++ {
							vs[i] = v.Index(i).Interface()
						}
						argsCodes = append(argsCodes, fmt.Sprintf("[]any{"+strings.Repeat(",%s", v.Len())[1:]+"}", vs...))
					}
				default:
					argsCodes = append(argsCodes, fmt.Sprintf("%v", a))
				}
			}
		}
		condCodes = append(condCodes, fmt.Sprintf("%s(%s)", cc[0], strings.Join(argsCodes, ", ")))
	}
	if len(c.useAlias) != 0 {
		if c.column.Alias() == c.column.OColName() {
			for _, ua := range c.useAlias {
				if ua.ColDef() == c.column.ColDef() {
					return fmt.Sprintf("%s.%s", strcase.ToCamel(c.column.Factory().CodeName()+" "+c.column.SrcName()), strings.Join(condCodes, "."))
				}
			}
			return fmt.Sprintf("%s.%s", c.column.SrcName(), strings.Join(condCodes, "."))
		} else {
			return fmt.Sprintf("%s.%s", c.column.Alias(), strings.Join(condCodes, "."))
		}
	} else {
		return fmt.Sprintf("%s.%s", c.column.SrcName(), strings.Join(condCodes, "."))
	}
}

func (c ConditionForBuilder[T]) ToCode() string {
	var (
		condCodes []string
	)
	for _, cc := range c.conditions {
		argsCodes := []string{}
	colLoop:
		for _, a := range cc[1:] {
			if col, ok := a.(preformShare.IColDef); ok {
				if len(c.useAlias) != 0 {
					if col.Alias() == col.OColName() {
						for _, ua := range c.useAlias {
							if ua.ColDef() == col.ColDef() {
								argsCodes = append(argsCodes, fmt.Sprintf("d.%sSchema.%s.%s", col.Factory().SchemaName(), col.Factory().CodeName(), strcase.ToCamel(col.Factory().CodeName()+" "+col.SrcName())))
								break colLoop
							}
						}
						argsCodes = append(argsCodes, fmt.Sprintf("d.%sSchema.%s.%s", col.Factory().SchemaName(), col.Factory().Alias(), col.SrcName()))
					} else {
						argsCodes = append(argsCodes, fmt.Sprintf("d.%sSchema.%s.%s", col.Factory().SchemaName(), col.Factory().Alias(), col.Alias()))
					}
				} else {
					argsCodes = append(argsCodes, fmt.Sprintf("d.%sSchema.%s.%s", col.Factory().SchemaName(), col.Factory().Alias(), col.SrcName()))
				}
			} else {
				t := reflect.TypeOf(a)
				switch t.Kind() {
				case reflect.String:
					argsCodes = append(argsCodes, fmt.Sprintf("\"%s\"", a))
				case reflect.Array, reflect.Slice:
					if cc[0] == "And" || cc[0] == "Or" {

						var (
							conds = a.([]preformShare.ICondForBuilder)
						)
						for _, cond := range conds {
							argsCodes = append(argsCodes, cond.ToCode())
						}
					} else {
						var (
							v  = reflect.ValueOf(a)
							vs = make([]any, v.Len())
						)
						for i := 0; i < v.Len(); i++ {
							vs[i] = v.Index(i).Interface()
						}
						argsCodes = append(argsCodes, fmt.Sprintf("[]any{"+strings.Repeat(",%s", v.Len())[1:]+"}", vs...))
					}
				default:
					argsCodes = append(argsCodes, fmt.Sprintf("%v", a))
				}
			}
		}
		condCodes = append(condCodes, fmt.Sprintf("%s(%s)", cc[0], strings.Join(argsCodes, ", ")))
	}
	if len(c.useAlias) != 0 {
		if c.column.Alias() == c.column.OColName() {
			for _, ua := range c.useAlias {
				if ua.ColDef() == c.column.ColDef() {
					return fmt.Sprintf("d.%sSchema.%s.%s.%s", c.column.Factory().SchemaName(), c.column.Factory().CodeName(), strcase.ToCamel(c.column.Factory().CodeName()+" "+c.column.SrcName()), strings.Join(condCodes, "."))
				}
			}
			return fmt.Sprintf("d.%sSchema.%s.%s.%s", c.column.Factory().SchemaName(), c.column.Factory().Alias(), c.column.SrcName(), strings.Join(condCodes, "."))
		} else {
			return fmt.Sprintf("d.%sSchema.%s.%s.%s", c.column.Factory().SchemaName(), c.column.Factory().Alias(), c.column.Alias(), strings.Join(condCodes, "."))
		}
	} else {
		return fmt.Sprintf("d.%sSchema.%s.%s.%s", c.column.Factory().SchemaName(), c.column.Factory().Alias(), c.column.SrcName(), strings.Join(condCodes, "."))
	}
}

// Eq
func (c ConditionForBuilder[T]) Eq(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Eq", value})
	return c
}

// Neq
func (c ConditionForBuilder[T]) NotEq(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"NotEq", value})
	return c
}

// Gt
func (c ConditionForBuilder[T]) Gt(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Gt", value})
	return c
}

// GtOrEq
func (c ConditionForBuilder[T]) GtOrEq(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"GtOrEq", value})
	return c
}

// Lt
func (c ConditionForBuilder[T]) Lt(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Lt", value})
	return c
}

// LtOrEq
func (c ConditionForBuilder[T]) LtOrEq(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"LtOrEq", value})
	return c
}

// Between
func (c ConditionForBuilder[T]) Between(value1, value2 any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Between", value1, value2})
	return c
}

func (c ConditionForBuilder[T]) Any(value any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Any", value})
	return c
}

// HasAny
func (c ConditionForBuilder[T]) HasAny(value ...any) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"ArrayHasAny", value})
	return c
}

func (c ConditionForBuilder[T]) IAnd(cond ...preformShare.ICondForBuilder) preformShare.ICondForBuilder {
	c.conditions = append(c.conditions, []any{"And", cond})
	return c
}

func (c ConditionForBuilder[T]) And(cond ...preformShare.ICondForBuilder) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"And", cond})
	return c
}

func (c ConditionForBuilder[T]) Or(cond ...preformShare.ICondForBuilder) ConditionForBuilder[T] {
	c.conditions = append(c.conditions, []any{"Or", cond})
	return c
}
