package mxdisk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type sysBlockInfo struct {
	devPath   string
	sysPath   string
	ro        int64
	removable int64
	size      int64
	slaves    []string
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

func fetchSysBlock(path string) map[string]sysBlockInfo {
	mp := make(map[string]sysBlockInfo)
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
			size, _ := readIntFromFile(syspath + "/size")
			slaves := readSysBlockSlaveInPath(syspath + "/slaves/") // for detect mdadm slaves

			mp[base] = sysBlockInfo{
				devPath:   base,
				sysPath:   syspath,
				ro:        ro,
				removable: removable,
				size:      size,
				slaves:    slaves,
			}
		}
		return err
	})
	fmt.Println(mp)
	return mp
}
