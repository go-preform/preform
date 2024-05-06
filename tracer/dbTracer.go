package preformTracer

import (
	"context"
	sql "database/sql"
	"fmt"
	"github.com/go-preform/preform"
	preformShare "github.com/go-preform/preform/share"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"strconv"
	"time"
)

func dummyTrace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	return ctx, func(err error) func(fetched bool, err error) {
		return func(bool, error) {}
	}
}

func dummyTraceExec(ctx context.Context, driver string, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	return ctx, func(res sql.Result, err error) {}
}

type chainTracer struct {
	tracers []preform.ITracer
}

func NewChainTracer(tracers ...preform.ITracer) *chainTracer {
	return &chainTracer{tracers: tracers}
}

func (l *chainTracer) Error(driver, msg string, err error) {
	for _, tracer := range l.tracers {
		tracer.Error(driver, msg, err)
	}
}

func (l *chainTracer) Trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	var (
		fns1 = make([]func(err error) func(fetched bool, err error), len(l.tracers))
	)
	for i, tracer := range l.tracers {
		ctx, fns1[i] = tracer.Trace(ctx, driver, query, txId, args...)
	}
	return ctx, func(err error) func(fetched bool, err error) {
		var (
			fns2 = make([]func(fetched bool, err error), len(l.tracers))
		)
		for i, fn := range fns1 {
			fns2[i] = fn(err)
		}
		return func(fetched bool, err error) {
			for _, fn := range fns2 {
				fn(fetched, err)
			}
		}
	}
}

func (l *chainTracer) TraceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	var (
		fns1 = make([]func(res sql.Result, err error), len(l.tracers))
	)
	for i, tracer := range l.tracers {
		ctx, fns1[i] = tracer.TraceExec(ctx, driver, query, txId, args...)
	}
	return ctx, func(res sql.Result, err error) {
		for _, fn := range fns1 {
			fn(res, err)
		}
	}
}

func (l *chainTracer) HealthLoop(ctx context.Context, db preform.DB) {
	for _, tracer := range l.tracers {
		go tracer.HealthLoop(ctx, db)
	}
}

func (l *chainTracer) SetLv(logLv preformShare.LogLv) {
	for _, tracer := range l.tracers {
		tracer.SetLv(logLv)
	}
}

type plainTracer struct {
	logLv          preformShare.LogLv
	healthCtx      context.Context
	healthDb       preform.DB
	trace          func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error))
	traceExec      func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error))
	healthInterval time.Duration
}

func NewPlainTracer(logLv preformShare.LogLv, healthInterval time.Duration) *plainTracer {
	t := &plainTracer{logLv: logLv, healthInterval: healthInterval}
	t.SetLv(logLv)
	return t
}

func (l *plainTracer) SetLv(logLv preformShare.LogLv) {
	if logLv&preformShare.LogLv_Read == preformShare.LogLv_Read {
		l.trace = l._trace
	} else {
		l.trace = dummyTrace
	}
	if logLv&preformShare.LogLv_Exec == preformShare.LogLv_Exec {
		l.traceExec = l._traceExec
	} else {
		l.traceExec = dummyTraceExec
	}
	l.logLv = logLv
}

func (l plainTracer) Error(driver, msg string, err error) {
	fmt.Printf("Preform Error %s %s %v\n", driver, msg, err)
}

func (l plainTracer) Trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	return l.trace(ctx, driver, query, txId, args...)
}

func (l plainTracer) TraceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	return l.traceExec(ctx, driver, query, txId, args...)
}

func (l plainTracer) _trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	fmt.Printf("Preform Query %s %s %s %s %v\n", driver, txId, id, query, args)
	t1 := time.Now()
	return ctx, func(err error) func(fetched bool, err error) {
		fmt.Printf("Preform Query finish %s %v %v\n", id, time.Now().Sub(t1), err)
		return func(fetched bool, err error) {
			if fetched {
				fmt.Printf("Preform Query finish fetching %s %v %v\n", id, time.Now().Sub(t1), err)
			}
		}
	}
}

func (l plainTracer) _traceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	fmt.Printf("Preform Exec %s %s %s %s %v\n", driver, txId, id, query, args)
	t1 := time.Now()
	return ctx, func(res sql.Result, err error) {
		var affected int64
		if err == nil {
			affected, err = res.RowsAffected()
		}
		fmt.Printf("Preform Exec finish %s %v affected: %v %v\n", id, time.Now().Sub(t1), affected, err)
	}
}

func (l plainTracer) HealthLoop(ctx context.Context, db preform.DB) {
	if l.logLv&preformShare.LogLv_Health == preformShare.LogLv_Health && l.healthInterval != 0 {
		var (
			ticker     = time.NewTicker(l.healthInterval)
			stats      sql.DBStats
			driverName = db.Db().DriverName()
			Db         = db.Db().DB
		)
		for {
			select {
			case <-ticker.C:
				stats = Db.Stats()
				fmt.Printf("Preform HealthLoop %s open: %v inUse: %v wait: %v %v\n", driverName, stats.OpenConnections, stats.InUse, stats.WaitCount, stats.WaitDuration)
			case <-ctx.Done():
				return
			}
		}
	}
}

type zeroLogTracer struct {
	logLv           preformShare.LogLv
	healthCtx       context.Context
	healthDb        preform.DB
	trace           func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error))
	traceExec       func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error))
	healthInterval  time.Duration
	logger          zerolog.Logger
	valueFromCtx    []string
	hasValueFromCtx bool
}

func NewZeroLogTracer(logger zerolog.Logger, logLv preformShare.LogLv, healthInterval time.Duration, valueFromCtx ...string) *zeroLogTracer {
	t := &zeroLogTracer{logLv: logLv, healthInterval: healthInterval, logger: logger, valueFromCtx: valueFromCtx, hasValueFromCtx: len(valueFromCtx) != 0}
	t.SetLv(logLv)
	return t
}

func (l *zeroLogTracer) SetLv(logLv preformShare.LogLv) {
	if logLv&preformShare.LogLv_Read == preformShare.LogLv_Read {
		l.trace = l._trace
	} else {
		l.trace = dummyTrace
	}
	if logLv&preformShare.LogLv_Exec == preformShare.LogLv_Exec {
		l.traceExec = l._traceExec
	} else {
		l.traceExec = dummyTraceExec
	}
	l.logLv = logLv
}

func (l zeroLogTracer) Error(driver, msg string, err error) {
	l.logger.Error().Str("driver", driver).Str("msg", msg).Err(err).Msg("Error")
}

func (l zeroLogTracer) Trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	return l.trace(ctx, driver, query, txId, args...)
}

func (l zeroLogTracer) TraceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	return l.traceExec(ctx, driver, query, txId, args...)
}

func (l zeroLogTracer) _trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	var (
		logger zerolog.Logger
	)
	if l.hasValueFromCtx {
		loggerCtx := l.logger.With().Str("preform", id).Str("driver", driver)
		for _, key := range l.valueFromCtx {
			loggerCtx = loggerCtx.Interface(key, ctx.Value(key))
		}
		logger = loggerCtx.Logger()
	} else {
		logger = l.logger.With().Str("preform", id).Str("driver", driver).Logger()
	}
	if txId != "" {
		logger = logger.With().Str("txId", txId).Logger()
	}
	logger.Debug().Str("query", query).Interface("args", args).Msg("Start")
	t1 := time.Now()
	return context.WithValue(ctx, preformShare.CTX_LOGGER, logger.Debug().Msg), func(err error) func(fetched bool, err error) {
		logger.Debug().Dur("cost", time.Now().Sub(t1)).Err(err).Msg("Finish Query")
		return func(fetched bool, err error) {
			if fetched {
				logger.Debug().Dur("time", time.Now().Sub(t1)).Err(err).Msg("Finish Fetching")
			}
		}
	}
}

func (l zeroLogTracer) _traceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	var (
		logger zerolog.Logger
	)
	if l.hasValueFromCtx {
		loggerCtx := l.logger.With().Str("preform", id).Str("driver", driver)
		for _, key := range l.valueFromCtx {
			loggerCtx = loggerCtx.Interface(key, ctx.Value(key))
		}
		logger = loggerCtx.Logger()
	} else {
		logger = l.logger.With().Str("preform", id).Str("driver", driver).Logger()
	}
	if txId != "" {
		logger = logger.With().Str("txId", txId).Logger()
	}
	logger.Debug().Str("exec", query).Interface("args", args).Msg("Start")
	t1 := time.Now()
	return context.WithValue(ctx, preformShare.CTX_LOGGER, logger.Debug().Msg), func(res sql.Result, err error) {
		var affected int64
		if err == nil {
			affected, err = res.RowsAffected()
		}
		logger.Debug().Dur("cost", time.Now().Sub(t1)).Err(err).Int64("affected", affected).Msg("Finish")
	}
}

func (l zeroLogTracer) HealthLoop(ctx context.Context, db preform.DB) {
	if l.logLv&preformShare.LogLv_Health == preformShare.LogLv_Health && l.healthInterval != 0 {
		var (
			ticker     = time.NewTicker(l.healthInterval)
			stats      sql.DBStats
			driverName = db.Db().DriverName()
			Db         = db.Db().DB
		)
		for {
			select {
			case <-ticker.C:
				stats = Db.Stats()
				l.logger.Info().Str("driver", driverName).Int("open", stats.OpenConnections).Int("inUse", stats.InUse).Int64("wait", stats.WaitCount).Dur("waitTime", stats.WaitDuration).Msg("HealthLoop")

			case <-ctx.Done():
				return
			}
		}
	}
}

type otelTracer struct {
	logLv           preformShare.LogLv
	healthCtx       context.Context
	healthDb        preform.DB
	trace           func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error))
	traceExec       func(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error))
	healthInterval  time.Duration
	tracer          trace.Tracer
	logErr          func(driver, msg string, err error)
	logStats        func(driver string, stats sql.DBStats)
	valueFromCtx    []string
	hasValueFromCtx bool
}

func NewOtelTracer(
	tracer trace.Tracer,
	logErr func(driver, msg string, err error),
	logStats func(driver string, stats sql.DBStats),
	logLv preformShare.LogLv,
	healthInterval time.Duration,
	valueFromCtx ...string,
) *otelTracer {
	t := &otelTracer{logLv: logLv, healthInterval: healthInterval, logErr: logErr, logStats: logStats, tracer: tracer, valueFromCtx: valueFromCtx, hasValueFromCtx: len(valueFromCtx) != 0}
	t.SetLv(logLv)
	return t
}

func (l *otelTracer) SetLv(logLv preformShare.LogLv) {
	if logLv&preformShare.LogLv_Read == preformShare.LogLv_Read {
		l.trace = l._trace
	} else {
		l.trace = dummyTrace
	}
	if logLv&preformShare.LogLv_Exec == preformShare.LogLv_Exec {
		l.traceExec = l._traceExec
	} else {
		l.traceExec = dummyTraceExec
	}
	l.logLv = logLv
}

func (l otelTracer) Error(driver, msg string, err error) {
	if l.logErr != nil {
		l.logErr(driver, msg, err)
	}
}

func (l otelTracer) Trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	return l.trace(ctx, driver, query, txId, args...)
}

func (l otelTracer) TraceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	return l.traceExec(ctx, driver, query, txId, args...)
}

func (l otelTracer) _trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	var (
		end   trace.Span
		attrs = []attribute.KeyValue{attribute.String("id", id), attribute.String("driver", driver), attribute.String("query", query), attribute.String("args", fmt.Sprintf("%v", args))}
	)
	if l.hasValueFromCtx {
		for _, key := range l.valueFromCtx {
			attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", ctx.Value(key))))
		}
	}
	if txId != "" {
		attrs = append(attrs, attribute.String("txId", txId))
	}
	ctx, end = l.tracer.Start(ctx, "Preform Query", trace.WithAttributes(attrs...))
	return context.WithValue(ctx, preformShare.CTX_LOGGER, func(msg string) {
			end.AddEvent(msg)
		}), func(err error) func(fetched bool, err error) {
			if err != nil {
				end.SetAttributes(attribute.String("error", err.Error()))
			}
			end.AddEvent("Finish Query")
			return func(fetched bool, err error) {
				if fetched {
					if err != nil {
						end.SetAttributes(attribute.String("error", err.Error()))
					}
				}
				end.End()
			}
		}
}

func (l otelTracer) _traceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error)) {
	id := strconv.FormatInt(rand.Int63(), 36)
	var (
		end   trace.Span
		attrs = []attribute.KeyValue{attribute.String("id", id), attribute.String("driver", driver), attribute.String("query", query), attribute.String("args", fmt.Sprintf("%v", args))}
	)
	if l.hasValueFromCtx {
		for _, key := range l.valueFromCtx {
			attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", ctx.Value(key))))
		}
	}
	if txId != "" {
		attrs = append(attrs, attribute.String("txId", txId))
	}
	ctx, end = l.tracer.Start(ctx, "Preform Exec", trace.WithAttributes(attrs...))
	return context.WithValue(ctx, preformShare.CTX_LOGGER, func(msg string) {
			end.AddEvent(msg)
		}), func(res sql.Result, err error) {
			var affected int64
			if err == nil {
				affected, err = res.RowsAffected()
			}
			if err != nil {
				end.SetAttributes(attribute.String("error", err.Error()))
			}
			end.SetAttributes(attribute.Int64("affected", affected))
			end.End()
		}
}

func (l otelTracer) HealthLoop(ctx context.Context, db preform.DB) {
	if l.logStats != nil && l.logLv&preformShare.LogLv_Health == preformShare.LogLv_Health && l.healthInterval != 0 {
		var (
			ticker     = time.NewTicker(l.healthInterval)
			stats      sql.DBStats
			driverName = db.Db().DriverName()
			Db         = db.Db().DB
		)
		for {
			select {
			case <-ticker.C:
				stats = Db.Stats()
				l.logStats(driverName, stats)
			case <-ctx.Done():
				return
			}
		}
	}
}
