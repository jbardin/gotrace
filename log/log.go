// Custom logger for gotrace
package log

import (
	"io"
	"log"
	"os"
	"sync/atomic"
)

// counter to mark each call so that entry and exit points can be correlated
var counter uint64

var L *log.Logger

// Setup our logger
// return  a value so this van be executed in a toplevel var statement
func Setup(output, prefix string) int {
	var out io.Writer
	switch output {
	case "stdout":
		out = os.Stdout
	default:
		out = os.Stderr
	}

	L = log.New(out, prefix, log.Lmicroseconds)
	return 0
}

func Next() uint64 {
	return atomic.AddUint64(&counter, 1)
}
