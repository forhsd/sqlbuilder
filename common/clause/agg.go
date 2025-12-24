package clause

type AggFunc string // count, sum, avg, max, min, distinct; 数据库的聚合函数不支持嵌套调用

const (
	Count  AggFunc = "count"
	Count1 AggFunc = "count1"
	Sum    AggFunc = "sum"
	Avg    AggFunc = "avg"
	Max    AggFunc = "max"
	Min    AggFunc = "min"
	// CountDistinct 不建议一个字符串拼接塞两个表达式的设计, 应该算作n个表达式归纳为自定义表达式的设计
	CountDistinct AggFunc = "countDistinct"
	// Stddev pg 总体标准差
	Stddev AggFunc = "stddev"
	// Variance pg 方差
	Variance AggFunc = "variance"

	PGMedian    AggFunc = "pg_median"
	DorisMedian AggFunc = "doris_median"

	PGCountIf         AggFunc = "countIf"
	PGSumIf           AggFunc = "sumIf"
	PGAvgIf           AggFunc = "avgIf"
	PGMaxIf           AggFunc = "maxIf"
	PGMinIf           AggFunc = "minIf"
	PGCountDistinctIf AggFunc = "countDistinctIf"
)

const (
	CountFormat         = `count(%s)`
	Count1Format        = `count(1)`
	SumFormat           = `sum(%s)`
	AvgFormat           = `avg(%s)`
	MaxFormat           = `max(%s)`
	MinFormat           = `min(%s)`
	CountDistinctFormat = `count(distinct(%s))`
	StddevFormat        = `stddev(%s)`
	VarianceFormat      = `variance(%s)`

	PGCountAsFormat         = `count(%s) as "%s"`
	PGCount1AsFormat        = `count(1) as "%s"`
	PGSumAsFormat           = `sum(%s) as "%s"`
	PGAvgAsFormat           = `avg(%s) as "%s"`
	PGMaxAsFormat           = `max(%s) as "%s"`
	PGMinAsFormat           = `min(%s) as "%s"`
	PGCountDistinctAsFormat = `count(distinct(%s)) as "%s"`
	PGStddevFormatAs        = `stddev(%s) as "%s"`
	PGVarianceFormatAs      = `variance(%s) as "%s"`

	PGCountIfFormat         = `count(%s) FILTER (WHERE %s)`
	PGSumIfFormat           = `sum(%s) FILTER (WHERE %s)`
	PGAvgIfFormat           = `avg(%s) FILTER (WHERE %s)`
	PGMaxIfFormat           = `max(%s) FILTER (WHERE %s)`
	PGMinIfFormat           = `min(%s) FILTER (WHERE %s)`
	PGCountDistinctIfFormat = `count(distinct(%s)) FILTER (WHERE %s)`
)

const (
	PGMedianFormat    = `percentile_cont(0.5) within group (order by %s)`
	DorisMedianFormat = `percentile_approx(%s, 0.5)`
)

// AggregationFuncFormatMap 映射聚合函数
var (
	AggregationFuncFormatMap = map[AggFunc]string{
		Count:         CountFormat,
		Count1:        Count1Format,
		Sum:           SumFormat,
		Avg:           AvgFormat,
		Max:           MaxFormat,
		Min:           MinFormat,
		CountDistinct: CountDistinctFormat,
		Stddev:        StddevFormat,
		Variance:      VarianceFormat,

		PGMedian:    PGMedianFormat,
		DorisMedian: DorisMedianFormat,

		PGSumIf:           PGSumIfFormat,
		PGAvgIf:           PGAvgIfFormat,
		PGMaxIf:           PGMaxIfFormat,
		PGMinIf:           PGMinIfFormat,
		PGCountIf:         PGCountIfFormat,
		PGCountDistinctIf: PGCountDistinctIfFormat,
	}

	PGAggregationAsFuncFormatMap = map[AggFunc]string{
		Count:         PGCountAsFormat,
		Count1:        PGCount1AsFormat,
		Sum:           PGSumAsFormat,
		Avg:           PGAvgAsFormat,
		Max:           PGMaxAsFormat,
		Min:           PGMinAsFormat,
		CountDistinct: PGCountDistinctAsFormat,
		Stddev:        PGStddevFormatAs,
		Variance:      PGVarianceFormatAs,
	}
)

type AggClause struct {
	// 聚合字段
	Aggregations []Expression `json:"aggregations"`
}
