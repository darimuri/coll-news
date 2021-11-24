package util

import (
	"fmt"
	"runtime/debug"
)

func PanicAsError(v interface{}) (retErr error) {
	if v == nil {
		return
	}

	switch t := v.(type) {
	case error:
		retErr = fmt.Errorf(fmt.Sprintf("panic '%s' from stack:\n%s", t.Error(), string(debug.Stack())))
	default:
		retErr = fmt.Errorf("errorless panic %+v from %s", v, string(debug.Stack()))
	}
	return
}
