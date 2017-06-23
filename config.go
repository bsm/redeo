package redeo

import "time"

// Config holds the server configuration
type Config struct {
	// Timeout represents the per-request socket read/write timeout.
	// Default: 0 (disabled)
	Timeout time.Duration

	// IdleTimeout forces servers to close idle connection once timeout is reached.
	// Default: 0 (disabled)
	IdleTimeout time.Duration

	// If non-zero, use SO_KEEPALIVE to send TCP ACKs to clients in absence
	// of communication. This is useful for two reasons:
	// 1) Detect dead peers.
	// 2) Take the connection alive from the point of view of network
	//    equipment in the middle.
	// On Linux, the specified value (in seconds) is the period used to send ACKs.
	// Note that to close the connection the double of the time is needed.
	// On other kernels the period depends on the kernel configuration.
	// Default: 0 (disabled)
	TCPKeepAlive time.Duration
}
