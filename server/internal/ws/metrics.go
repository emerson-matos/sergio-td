package ws

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	metricTickCount        atomic.Int64
	metricTickDurationSum  atomic.Int64
	metricTickDurationMax  atomic.Int64
	metricCommandsReceived atomic.Int64
	metricCommandsRejected atomic.Int64
	metricActiveConns      atomic.Int64
)

func RecordTick(duration time.Duration) {
	metricTickCount.Add(1)
	us := duration.Microseconds()
	metricTickDurationSum.Add(us)
	for {
		cur := metricTickDurationMax.Load()
		if us <= cur || metricTickDurationMax.CompareAndSwap(cur, us) {
			break
		}
	}
}

func Metrics(w http.ResponseWriter, _ *http.Request) {
	ticks := metricTickCount.Load()
	sumUs := metricTickDurationSum.Load()
	maxUs := metricTickDurationMax.Load()

	var avgUs int64
	if ticks > 0 {
		avgUs = sumUs / ticks
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ticks":             ticks,
		"tickDurationAvgUs": avgUs,
		"tickDurationMaxUs": maxUs,
		"commandsReceived":  metricCommandsReceived.Load(),
		"commandsRejected":  metricCommandsRejected.Load(),
		"activeConnections": metricActiveConns.Load(),
	})
}
