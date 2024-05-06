package dialect

import preformShare "github.com/go-preform/preform/share"

const (
	AggSum           preformShare.Aggregator = "SUM"
	AggAvg           preformShare.Aggregator = "AVG"
	AggMax           preformShare.Aggregator = "MAX"
	AggMin           preformShare.Aggregator = "MIN"
	AggCount         preformShare.Aggregator = "COUNT"
	AggCountDistinct preformShare.Aggregator = "COUNT_DISTINCT"
	AggMean          preformShare.Aggregator = "MEAN"
	AggMedian        preformShare.Aggregator = "MEDIAN"
	AggMode          preformShare.Aggregator = "MODE"
	AggStdDev        preformShare.Aggregator = "STDDEV"
	AggGroupConcat   preformShare.Aggregator = "GROUP_CONCAT"
	AggJson          preformShare.Aggregator = "JSON_AGG"
	AggArray         preformShare.Aggregator = "ARRAY_AGG"
)
