// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	_ "github.com/burik666/yagostatus/widgets"
	"github.com/burik666/yagostatus/ygs"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.Ldate + log.Ltime + log.Lshortfile)

	configFile := flag.String("config", "yagostatus.yml", "Config file")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	yaGoStatus := YaGoStatus{}

	for _, w := range config.Widgets {
		widget, ok := ygs.NewWidget(w.Name + "widget")
		if !ok {
			log.Fatalf("Widget '%s' not found", w.Name)
		}

		err := yaGoStatus.AddWidget(widget, w)
		if err != nil {
			log.Fatalf("Widget '%s' configuration error: %s", w.Name, err)
		}
	}

	stopsignals := make(chan os.Signal, 1)
	signal.Notify(stopsignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		yaGoStatus.Run()
		stopsignals <- syscall.SIGTERM
	}()
	<-stopsignals
	yaGoStatus.Stop()
}
