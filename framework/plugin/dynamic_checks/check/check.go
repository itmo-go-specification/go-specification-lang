package check

import (
	"fmt"
	"runtime"
)

func check(condition bool, messages ...any) {
	if !condition {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "unknown"
			line = 0
		}

		msg := fmt.Sprintf("check failed at %s:%d", file, line)
		if len(messages) > 0 {
			msg += ": " + fmt.Sprint(messages...)
		}

		panic(msg)
	}
}
