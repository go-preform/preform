# Preform
### A productive ORM in GO

[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

Compile data models down to column level aim at querying without using any string, 
by knowing the data types scanning data can be faster than hand-writing rows.Scan even it's still the trusted official drivers.

go 1.18+, currently supports Postgres, Mysql, Clickhouse, Sqlite and more is coming.

## Overview

- **Model builder**, will build table structure for queries & body type as data container
- **Relations**, eager loading, join with predefined foreign keys
- **Query builder** base on Masterminds/squirrel with type safe output
- **Prebuild complex queries** for fast querying & type safe output
- **Condition without string**, prebuild columns definition in table structure to avoid using string
- **Type specific scanner**, make it faster than native rows.Scan
- **Production grade Performance**, avoid using reflect after initialization
- **Schema wrapping**, easy to archive schema data isolation 
- **Flexible log and tracing**, easy integration with zerolog, otel, etc.
- **AI friendly**, pre-generated code is easy for AI to understand compare to string

## Example

Talk is cheap, show me the code!

#### Quering
```go
model.MainSchema.Use(func(mainSchema *model.MainSchema) { // syntax sugar to avoid long prefix chain
  // simple select
  user, err := mainSchema.ModelWithMultipleKeys.Select(). // select all columns
    GetOne("keyValue1", "keyValue2"...)                   // get by primary key value(s)

  // select with eager loading
  user, err := mainSchema.User.Select(mainSchema.User.Id, mainSchema.User.Username). // select columns with predefined column fields
    Where(mainSchema.User.Id.Eq(191234).And(mainSchema.User.Flags.Any(1))).          // condition without hand-writing string
    Eager(mainSchema.User.Orders.Columns(mainSchema.Order.Amount).                   // eager loading with options
        Eager(mainSchema.Order.Products)).                                           // multiple levels eager loading
    Ctx(ctx).GetOne()
  
  // relation loading
  cards, err := user.LoadCards() 
  

  // join query
  rows, err := mainSchema.User.Select(mainSchema.User.Id, mainSchema.User.Username, mainSchema.User.Bookmarks, mainSchema.UserBookmark.Name).
    JoinForeignKey(mainSchema.User.BookmarkIds). // join with predefined foreign key
    Query()
})
```

#### CUD
```go
// insert
err = mainSchema.User.Insert(&model.UserBody{...}, preform.EditConfig{Tx: tx, Ctx: ctx}) // with optional insert config

err = mainSchema.User.InsertBatch([]*model.UserBody{...})

err = user.Insert(preform.EditConfig{Cascading: true})

// update
err = user.Update(preform.UpdateConfig{Tx: tx, Ctx: ctx, Cols: []preform.ICol{mainSchema.User.Username}})  // with optional update config

affected, err = mainSchema.User.Update().Set(mainSchema.User.Username, "test").Where(mainSchema.User.Id.Eq(1)).Exec()

// delete
deleted, err = user.Delete(preform.EditConfig{Tx: tx, Ctx: ctx})  // with optional delete config

deleted, err = mainSchema.User.Delete().Where(mainSchema.User.Id.Eq(1)).Exec()
```

#### Building models
```go
// build with main.go, using code to generate is more straightforward & flexible than cli IMO
preform.BuildModel(nativeDbConn, "pkgName", "outputPath", "schema1", "schema2" ...) 
```

#### Customize model
```go
// models will be generated along with source code in src folder, add go file to define more advanced structure
func (d *MainSchema_order) Setup() (skipAutoSetter bool) {
  d.Status.OverwriteType(preform.ColumnDef[CUSTOM_STATUS_ENUM]{})                // overwrite column type
  d.UserId.SetAssociatedKey(MainSchema.user.Id, preform.FkRelationName("Buyer")) // custom foreign key field, set relation name in this case
  d.CardId.SetAssociatedKey(MainSchema.card.Id)                                  // retain auto joining from original generated code
  return true
}
```

#### Prebuild queries
```go
// define in src folder init() as part of customize model
PrebuildQueries = append(PrebuildQueries, queryBuilder.Build("getNotificationCount",
  func(builder *queryBuilder.QueryBuilder, main *MainSchema) {
    builder.From(main.user).
      Cols(
        main.user.Id.SetAlias("UId"),                   // field alias
        main.Notification.Id.Count().SetAlias("Cnt"),   // field alias
        main.Notification.Priority,
      ).
    LeftJoin(main.Notification, main.Notification.TargetIds.Any(main.user.Id).And(main.Notification.Target.Eq(2))).  // joining condition
    Where(main.Notification.Target.Eq(2)).                                                                           // predefine condition
    Having(main.Notification.Status.NotEq(0)).                                                                       // predefine having
    GroupBy(main.user.Id, main.Notification.Priority)                                                                // predefine group by
}))

// query
notes, err := model.GetAdminNotifications.Select(model.GetNotificationCount.Cnt, model.GetNotificationCount.UId). // custom select columns
    Where(model.GetNotificationCount.UId.Eq(1)).                                                                  // additional where condition
    GetAll() // type safe output
```

### Benchmarks
```bash
goos: windows
goarch: amd64
pkg: github.com/go-preform/preform/benchmark
cpu: 12th Gen Intel(R) Core(TM) i7-1280P
BenchmarkPreformSelectAll-20                 472           2608712 ns/op          314256 B/op      12235 allocs/op
BenchmarkPreformSelectAllFast-20             526           2367094 ns/op          260900 B/op       9341 allocs/op
BenchmarkPreformSelectEager-20               133           8811348 ns/op         2112758 B/op      31962 allocs/op
BenchmarkPreformSelectEagerFast-20           142           8522567 ns/op         2001407 B/op      25982 allocs/op
BenchmarkGormSelectAll-20                    326           3721617 ns/op          416996 B/op      23210 allocs/op
BenchmarkGormSelectEager-20                   91          13053397 ns/op         3532487 B/op      67214 allocs/op
BenchmarkEntSelectAll-20                     423           2910246 ns/op          639377 B/op      20357 allocs/op
BenchmarkEntSelectEager-20                   132           9301639 ns/op         2027165 B/op      48898 allocs/op
BenchmarkSqlxStructScan-20                   409           2961220 ns/op          651155 B/op      12183 allocs/op
BenchmarkSqlRawScan-20                       432           2836293 ns/op          650705 B/op      12180 allocs/op
```


