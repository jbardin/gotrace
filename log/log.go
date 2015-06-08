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

// Return a new logger.
func New(output, prefix string) *log.Logger {
	var out io.Writer
	switch output {
	case "stdout":
		out = os.Stdout
	default:
		out = os.Stderr
	}

	return log.New(out, prefix, log.Lmicroseconds)
}

func Next() uint64 {
	return atomic.AddUint64(&counter, 1)
}
