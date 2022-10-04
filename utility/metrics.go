package utility

import (
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var MCounter = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:        "func_time_cost_seconds",
	ConstLabels: map[string]string{},
	Buckets:     []float64{},
}, []string{"func_name"})

func init() {
	// 启动监控协程
	go Detector()
}

// Detector 检测器
func Detector() {

}

// GetTimeCost  获取耗时
//
//	@param startTime
func GetTimeCost(startTime time.Time, function string) {
	cost := time.Now().Sub(startTime).Seconds()
	MCounter.WithLabelValues(function).Observe(cost)
}

func RunFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	a := strings.Split(f.Name(), "/")
	return a[len(a)-1]
}
