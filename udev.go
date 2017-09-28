package mxdisk

// UdevInfo info from /dev/disk/by-xxx
type UdevInfo struct {
	UUID  string
	Label string
	Path  string
}
