package internal

import (
	"fmt"
	"path"
	"runtime"
)

var Root string

func init() {
	_, filename, _, ok := runtime.Caller(0) // Get information about the current caller (this file)
	if !ok {
		fmt.Println("Unable to get the current filename.")
		return
	}
	Root = path.Dir(path.Dir(path.Dir(filename)))
}
