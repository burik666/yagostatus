// Package ygs contains the YaGoStatus structures.
package ygs

// Widget represents a widget struct.
type Widget interface {
	Run(chan []I3BarBlock) error
	Event(I3BarClickEvent)
	Stop()
	Configure(map[string]interface{}) error
}
