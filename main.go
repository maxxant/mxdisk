package main
import (
	"fmt"
	"github.com/maxxant/go-fstab"
)

func main() {
	if mn, err := fstab.ParseFile("/proc/mounts"); err != nil {
		panic("aaa")
	} else {
		fmt.Println(mn)
	}
}