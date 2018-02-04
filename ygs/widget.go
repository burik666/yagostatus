package ygs

type Widget interface {
	Run(chan []I3BarBlock) error
	Event(I3BarClickEvent)
	Stop()
	Configure(map[string]interface{}) error
}
