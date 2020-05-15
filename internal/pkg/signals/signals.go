package signals

// #include <signal.h>
import "C"

// SIGRTMIN signal
var SIGRTMIN int

// SIGRTMAX signal
var SIGRTMAX int

func init() {
	SIGRTMIN = int(C.int(C.SIGRTMIN))
	SIGRTMAX = int(C.int(C.SIGRTMAX))
}
