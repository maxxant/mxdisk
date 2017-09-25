package mxdisk

const (
	// ConstMonitoringProcmountSec defaults config monitoring tick "/proc/mounts" in seconds
	ConstMonitoringProcmountSec = 2

	// ConstMonitoringFstabSec defaults config monitoring tick "/etc/fstab" in seconds
	ConstMonitoringFstabSec = 20

	// ConstOnlyUUIDMountedDisks defaults config for filtering /dev/loop & etc vfs
	ConstOnlyUUIDMountedDisks = true
)

// Config struct for operations
type Config struct {
	// MonitoringProcmountSec monitoring tick "/proc/mounts" in seconds, 1..N
	// defaults value: ConstMonitoringProcmountSec
	MonitoringProcmountSec int

	// MonitoringFstabSec monitoring tick "/etc/fstab" in seconds
	// full reload fstab without inotify & etc for vfs independence
	// range 1..N, or = 0 is disabled
	// defaults value: ConstMonitoringFstabSec
	MonitoringFstabSec int

	// OnlyUUIDMountedDisks is true for filtering /dev/loop & etc vfs (mounted disks & partitions)
	// defaults value: ConstOnlyUUIDMountedDisks
	OnlyUUIDMountedDisks bool
}

// NewConfig make default Config
func NewConfig() *Config {
	return &Config{
		MonitoringProcmountSec: ConstMonitoringProcmountSec,
		MonitoringFstabSec:     ConstMonitoringFstabSec,
		OnlyUUIDMountedDisks:   ConstOnlyUUIDMountedDisks,
	}
}
