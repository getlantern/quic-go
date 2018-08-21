package congestion

import (
	"time"

	"github.com/getlantern/quic-go/internal/protocol"
)

// A SendAlgorithm performs congestion control and calculates the congestion window
type SendAlgorithm interface {
	TimeUntilSend(bytesInFlight protocol.ByteCount) time.Duration
	OnPacketSent(sentTime time.Time, bytesInFlight protocol.ByteCount, packetNumber protocol.PacketNumber, bytes protocol.ByteCount, isRetransmittable bool)
	GetCongestionWindow() protocol.ByteCount
	MaybeExitSlowStart()
	OnPacketAcked(number protocol.PacketNumber, ackedBytes protocol.ByteCount, priorInFlight protocol.ByteCount, eventTime time.Time, leastUnacked protocol.PacketNumber)
	OnPacketLost(number protocol.PacketNumber, lostBytes protocol.ByteCount, priorInFlight protocol.ByteCount, eventTime time.Time, leastUnacked protocol.PacketNumber)
	SetNumEmulatedConnections(n int)
	OnRetransmissionTimeout(packetsRetransmitted bool)
	OnConnectionMigration()

	// Experiments
	SetSlowStartLargeReduction(enabled bool)
	BandwidthEstimate() Bandwidth
}

// SendAlgorithmWithDebugInfo adds some debug functions to SendAlgorithm
type SendAlgorithmWithDebugInfo interface {
	SendAlgorithm

	// Stuff only used in testing
	HybridSlowStart() *HybridSlowStart
	SlowstartThreshold() protocol.ByteCount
	RenoBeta() float32
	InRecovery() bool
}
