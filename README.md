# Preform
### A productive ORM in GO

[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

Compile data models down to column level aim at querying without using any string, 
by knowing the data types scanning data can be faster than hand-writing rows.Scan even it's still the trusted official drivers.

go 1.18+, currently supports Postgres, Mysql, Clickhouse, Sqlite and more is coming.

## Overview

- **Model builder**, will build table structure for queries & body type as data container, with custom type & enum supports
- **Relations**, eager loading, join with predefined foreign keys / middle table
- **Integrated query builder** base on [Masterminds/squirrel](http://github.com/Masterminds/squirrel) with type safe output
- **Prebuild complex queries** for fast querying & type safe output
- **Condition without string**, prebuild columns definition in table structure to avoid using string
- **Type specific scanner**, make it faster than native rows.Scan
- **Production grade Performance**, avoid using reflect after initialization and other optimization. Please see [benchmarks](#benchmarks) / [details](https://github.com/go-preform/preform/blob/pages/docs/whyFast.md)
- **Schema wrapping**, easy to archive schema data isolation 
- **Flexible log and tracing**, built in support with zerolog, otel, interface for custom logger/tracer
- **AI friendly**, pre-generated code is easy for AI to understand compare to string

![Flow chart](https://go-preform.github.io/preform/asset/preformFlow.png)
## Example

Talk is cheap, show me the code!

For more examples, please check [test](./test) folder

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
# with local docker postgres, 1000 rows on Apple M1 MAX
goos: darwin
goarch: arm64
pkg: github.com/go-preform/preform/benchmark
BenchmarkPreformSelectAll-20               1107    1114150 ns/op   313897 B/op    12231 allocs/op #1000 rows
BenchmarkPreformSelectAllFast-20           1598     718653 ns/op   260551 B/op     9337 allocs/op #1000 rows
BenchmarkPreformSelectEager-20              327    4195277 ns/op  2111897 B/op    31951 allocs/op #100 rows + 1000 + 1000
BenchmarkPreformSelectEagerFast-20          385    3184728 ns/op  2001004 B/op    25971 allocs/op #100 rows + 1000 + 1000
BenchmarkGormSelectAll-20                   506    2805352 ns/op   416675 B/op    23210 allocs/op #1000 rows
BenchmarkGormSelectEager-20                 129    9838620 ns/op  3537747 B/op    67227 allocs/op #100 rows + 1000 + 1000
BenchmarkEntSelectAll-20                    721    1592649 ns/op   639373 B/op    20357 allocs/op #1000 rows
BenchmarkEntSelectEager-20                  267    5305585 ns/op  2025764 B/op    48888 allocs/op #100 rows + 1000 + 1000
BenchmarkSqlxStructScan-20                  814    1529897 ns/op   651149 B/op    12183 allocs/op #1000 rows
BenchmarkSqlRawScan-20                      874    1350841 ns/op   650702 B/op    12180 allocs/op #1000 rows
```

[![Benchmarks](https://go-preform.github.io/preform/asset/benchChart.png)](https://go-preform.github.io/preform/asset/benchChart.png)
