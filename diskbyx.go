package mxdisk

import (
	"fmt"
	"os"
	"path/filepath"
)

// path:
//  /dev/disk/by-uuid
//  /dev/disk/by-label
//  /dev/disk/by-path
// returns map [by-xxx] /dev/sdxN
// NOTE: not all OS supports path "by-label"
func disksByPathX(path string) map[string]string {
	mp := make(map[string]string)
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if err != nil {
			return err // if path is not exists
		}
		if inf.IsDir() {
			return nil
		}
		if (inf.Mode() & os.ModeSymlink) != 0 {
			link := filepath.Base(path)
			name, _ := filepath.EvalSymlinks(path)
			mp[name] = link
		}
		return err
	})
	return mp
}

const (
	byUUID  = iota
	byLabel = iota
	byPath  = iota
)

type disksByX struct {
	// key:val -> /dev/sda1 : uuid
	uuid  map[string]string
	label map[string]string
	path  map[string]string
}

func (p disksByX) getMap(byXFilter int) map[string]string {
	switch byXFilter {
	case byUUID:
		return p.uuid
	case byLabel:
		return p.label
	case byPath:
		return p.path
	}
	panic("undefined filter index")
}

func (p disksByX) find(byXFilter int, need string) string {
	mp := p.getMap(byXFilter)
	if v, ok := mp[need]; ok {
		return v
	}
	return ""
}

func (p disksByX) String() string {
	var s string
	for k, v := range p.uuid {
		s += fmt.Sprintf("uuid: %v:%v\n", k, v)
	}
	for k, v := range p.label {
		s += fmt.Sprintf("label: %v:%v\n", k, v)
	}
	for k, v := range p.path {
		s += fmt.Sprintf("path: %v:%v\n", k, v)
	}
	return s
}

func newDisksByX() *disksByX {
	return &disksByX{
		uuid:  disksByPathX("/dev/disk/by-uuid"),
		label: disksByPathX("/dev/disk/by-label"),
		path:  disksByPathX("/dev/disk/by-path"),
	}
}
