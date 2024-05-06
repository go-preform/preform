package cachedQueryRunner

import (
	"database/sql/driver"
	preformShare "github.com/go-preform/preform/share"
	"github.com/maypok86/otter"
	"time"
)

type defaultCacher struct {
	otter.Cache[string, [][]driver.Value]
	queriesByFactory otter.Cache[string, *[]string]
}

func NewDefaultCacher(ttl time.Duration) *defaultCacher {
	c := &defaultCacher{}
	c.Cache, _ = otter.MustBuilder[string, [][]driver.Value](1000000).
		Cost(func(key string, value [][]driver.Value) uint32 {
			return uint32(len(value))
		}).
		WithTTL(ttl).
		Build()
	c.queriesByFactory, _ = otter.MustBuilder[string, *[]string](10000).
		WithTTL(ttl).
		Build()
	return c
}

func (d *defaultCacher) Load(key string) (value [][]driver.Value, saver func(value [][]driver.Value, relatedFactories []preformShare.IQueryFactory), ok bool) {
	v, ok := d.Get(key)
	if ok {
		return v, d.save(key), ok
	}
	return nil, d.save(key), ok
}

func (d *defaultCacher) save(key string) func(value [][]driver.Value, relatedFactories []preformShare.IQueryFactory) {
	return func(value [][]driver.Value, relatedFactories []preformShare.IQueryFactory) {
		d.Set(key, value)
		var (
			name    string
			queries *[]string
			ok      bool
		)
		for _, f := range relatedFactories {
			for _, name = range f.TableNames() {
				if queries, ok = d.queriesByFactory.Get(name); ok {
					*queries = append(*queries, key)
				} else {
					d.queriesByFactory.Set(name, &[]string{key})
				}
			}
		}
	}
}

func (d *defaultCacher) ClearByFactories(factories []preformShare.IQueryFactory) {
	var (
		name    string
		queries *[]string
		ok      bool
	)
	for _, f := range factories {
		for _, name = range f.TableNames() {
			if queries, ok = d.queriesByFactory.Get(name); ok {
				for _, query := range *queries {
					d.Delete(query)
				}
				d.queriesByFactory.Delete(name)
			}
		}
	}
}
