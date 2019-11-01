// Package ygs contains the YaGoStatus structures.
package ygs

// Widget represents a widget struct.
type Widget interface {
	Run(chan<- []I3BarBlock) error
	Event(I3BarClickEvent, []I3BarBlock)
	Stop()
	Continue()
	Shutdown()
}
