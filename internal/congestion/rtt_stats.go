package congestion

import (
	"sync/atomic"
	"time"

	"github.com/lucas-clemente/quic-go/internal/utils"
)

const (
	rttAlpha      float32 = 0.125
	oneMinusAlpha float32 = (1 - rttAlpha)
	rttBeta       float32 = 0.25
	oneMinusBeta  float32 = (1 - rttBeta)
	// The default RTT used before an RTT sample is taken.
	defaultInitialRTT = 100 * time.Millisecond
)

// RTTStats provides round-trip statistics
type RTTStats struct {
	minRTT        uint64 // time.Duration
	latestRTT     uint64 // time.Duration
	smoothedRTT   uint64 // time.Duration
	meanDeviation uint64 // time.Duration
}

// NewRTTStats makes a properly initialized RTTStats object
func NewRTTStats() *RTTStats {
	return &RTTStats{}
}

// MinRTT Returns the minRTT for the entire connection.
// May return Zero if no valid updates have occurred.
func (r *RTTStats) MinRTT() time.Duration { return time.Duration(atomic.LoadUint64(&r.minRTT)) }

// LatestRTT returns the most recent rtt measurement.
// May return Zero if no valid updates have occurred.
func (r *RTTStats) LatestRTT() time.Duration { return time.Duration(atomic.LoadUint64(&r.latestRTT)) }

// SmoothedRTT returns the EWMA smoothed RTT for the connection.
// May return Zero if no valid updates have occurred.
func (r *RTTStats) SmoothedRTT() time.Duration {
	return time.Duration(atomic.LoadUint64(&r.smoothedRTT))
}

// SmoothedOrInitialRTT returns the EWMA smoothed RTT for the connection.
// If no valid updates have occurred, it returns the initial RTT.
func (r *RTTStats) SmoothedOrInitialRTT() time.Duration {
	d := atomic.LoadUint64(&r.smoothedRTT)
	if d != 0 {
		return time.Duration(d)
	}
	return defaultInitialRTT
}

// MeanDeviation gets the mean deviation
func (r *RTTStats) MeanDeviation() time.Duration {
	return time.Duration(atomic.LoadUint64(&r.meanDeviation))
}

// UpdateRTT updates the RTT based on a new sample.
func (r *RTTStats) UpdateRTT(sendDelta, ackDelay time.Duration, now time.Time) {
	if sendDelta == utils.InfDuration || sendDelta <= 0 {
		return
	}

	// Update r.minRTT first. r.minRTT does not use an rttSample corrected for
	// ackDelay but the raw observed sendDelta, since poor clock granularity at
	// the client may cause a high ackDelay to result in underestimation of the
	// r.minRTT.
	minRTT := time.Duration(atomic.LoadUint64(&r.minRTT))
	if minRTT == 0 || minRTT > sendDelta {
		minRTT = sendDelta
		atomic.StoreUint64(&r.minRTT, uint64(minRTT))
	}

	// Correct for ackDelay if information received from the peer results in a
	// an RTT sample at least as large as minRTT. Otherwise, only use the
	// sendDelta.
	sample := sendDelta
	if sample-minRTT >= ackDelay {
		sample -= ackDelay
	}
	atomic.StoreUint64(&r.latestRTT, uint64(sample))
	// First time call.
	smoothedRTT := time.Duration(atomic.LoadUint64(&r.smoothedRTT))
	if smoothedRTT == 0 {
		atomic.StoreUint64(&r.smoothedRTT, uint64(sample))
		atomic.StoreUint64(&r.meanDeviation, uint64(sample/2))
	} else {
		meanDeviation := time.Duration(atomic.LoadUint64(&r.meanDeviation))
		atomic.StoreUint64(&r.meanDeviation, uint64(time.Duration(oneMinusBeta*float32(meanDeviation/time.Microsecond)+rttBeta*float32(utils.AbsDuration(smoothedRTT-sample)/time.Microsecond))*time.Microsecond))
		atomic.StoreUint64(&r.smoothedRTT, uint64(time.Duration((float32(smoothedRTT/time.Microsecond)*oneMinusAlpha)+(float32(sample/time.Microsecond)*rttAlpha))*time.Microsecond))
	}
}

// OnConnectionMigration is called when connection migrates and rtt measurement needs to be reset.
func (r *RTTStats) OnConnectionMigration() {
	atomic.StoreUint64(&r.latestRTT, 0)
	atomic.StoreUint64(&r.minRTT, 0)
	atomic.StoreUint64(&r.smoothedRTT, 0)
	atomic.StoreUint64(&r.meanDeviation, 0)
}

// ExpireSmoothedMetrics causes the smoothed_rtt to be increased to the latest_rtt if the latest_rtt
// is larger. The mean deviation is increased to the most recent deviation if
// it's larger.
func (r *RTTStats) ExpireSmoothedMetrics() {
	meanDeviation := time.Duration(atomic.LoadUint64(&r.meanDeviation))
	smoothedRTT := time.Duration(atomic.LoadUint64(&r.smoothedRTT))
	latestRTT := time.Duration(atomic.LoadUint64(&r.latestRTT))
	atomic.StoreUint64(&r.meanDeviation, uint64(utils.MaxDuration(meanDeviation, utils.AbsDuration(smoothedRTT-latestRTT))))
	atomic.StoreUint64(&r.smoothedRTT, uint64(utils.MaxDuration(smoothedRTT, latestRTT)))
}
