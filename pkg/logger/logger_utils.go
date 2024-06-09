package logger

func unsafeGetArgs(args ...interface{}) (resStr string, resArgs []interface{}) {

	switch t := args[0].(type) {
	case string:
		resStr = t
		if len(args) >= 2 {
			resArgs = args[1:]
		}
	case error:
		resStr = t.Error()
	}

	return
}
