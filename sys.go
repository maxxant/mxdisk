package mxdisk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// SysBlockInfo the info from /sys/block
type SysBlockInfo struct {
	Ro        int64
	Removable int64
	slaves    []string
}

// SysMapBlocks the map of /sys/block devices
type SysMapBlocks map[string]SysBlockInfo

func (p SysMapBlocks) String() string {
	var s string
	for _, v := range p {
		s += fmt.Sprintf("%+v\n", v)
	}
	return s
}

func (p SysMapBlocks) exposeDevsSlaves(devs []string) []string {
	mp := make(map[string]bool)
	for _, v := range devs {
		mp[v] = true
		if m, ok := p[v]; ok {
			for _, s := range m.slaves {
				mp[s] = true
			}
		}
	}

	var res []string
	for k := range mp {
		res = append(res, k)
	}

	// recursion for detect sub-sub slaves
	if len(res) > len(devs) {
		return p.exposeDevsSlaves(res)
	}
	return res
}

func readIntFromFile(f string) (val int64, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(f); err != nil {
		return
	}

	sdata := string(data[:len(data)-1])
	if val, err = strconv.ParseInt(sdata, 10, 0); err != nil {
		return
	}
	return val, err
}

func readSysBlockSlaveInPath(path string) []string {
	var res []string
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if err != nil {
			return err // if path is not exists
		}
		if inf.IsDir() {
			return nil
		}
		if (inf.Mode() & os.ModeSymlink) != 0 {
			base := "/dev/" + filepath.Base(path)
			res = append(res, base)
		}
		return err
	})
	return res
}

func fetchSysBlock(path string) SysMapBlocks {
	mp := make(SysMapBlocks)
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if err != nil {
			return err // if path is not exists
		}
		if inf.IsDir() {
			return nil
		}
		if (inf.Mode() & os.ModeSymlink) != 0 {
			base := "/dev/" + filepath.Base(path)
			syspath, _ := filepath.EvalSymlinks(path)
			ro, _ := readIntFromFile(syspath + "/ro")
			removable, _ := readIntFromFile(syspath + "/removable")
			slaves := readSysBlockSlaveInPath(syspath + "/slaves/") // for detect mdadm slaves

			mp[base] = SysBlockInfo{
				Ro:        ro,
				Removable: removable,
				slaves:    slaves,
			}
		}
		return err
	})
	return mp
}
