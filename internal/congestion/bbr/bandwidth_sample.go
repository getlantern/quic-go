package bbr

import (
	"time"

	"github.com/getlantern/quic-go/internal/congestion"
)

type bandwidthSample struct {
	// The bandwidth at that particular sample. Zero if no valid bandwidth sample
	// is available.
	bandwidth congestion.Bandwidth
	// The RTT measurement at this particular sample.  Zero if no RTT sample is
	// available.  Does not correct for delayed ack time.
	rtt time.Duration
	// Indicates whether the sample might be artificially low because the sender
	// did not have enough data to send in order to saturate the link.
	isAppLimited bool
}
