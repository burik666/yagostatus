// +build cgo

package signals

// #include <signal.h>
import "C"

// SIGRTMIN signal
var SIGRTMIN int = int(C.int(C.SIGRTMIN))

// SIGRTMAX signal
var SIGRTMAX int = int(C.int(C.SIGRTMAX))
