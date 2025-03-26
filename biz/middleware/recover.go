package middleware

import (
	"runtime"
)

const defaultStackSize = 4096

func GetCurrentGoroutineStack() string {
	var buf [defaultStackSize]byte
	size := runtime.Stack(buf[:], false)
	return string(buf[:size])
}

func RecoverHandler(p any) (err error) {
	log.Errorf("grpc panic: %v\nstack: %s\n", p, GetCurrentGoroutineStack())
	return errors.InternalServiceError
}
