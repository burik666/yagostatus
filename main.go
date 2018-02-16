// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	"fmt"
	_ "github.com/burik666/yagostatus/widgets"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.Ldate + log.Ltime + log.Lshortfile)

	configFile := flag.String("config", "yagostatus.yml", "config file")
	versionFlag := flag.Bool("version", false, "print version information and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("YaGoStatus %s\n", Version)
		return
	}

	yaGoStatus := YaGoStatus{}
	yaGoStatus.Configure(*configFile)

	stopsignals := make(chan os.Signal, 1)
	signal.Notify(stopsignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		yaGoStatus.Run()
		stopsignals <- syscall.SIGTERM
	}()
	<-stopsignals
	yaGoStatus.Stop()
}
