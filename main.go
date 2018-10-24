// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/burik666/yagostatus/widgets"
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
	err := yaGoStatus.Configure(*configFile)
	if err != nil {
		log.Fatalf("configure failed: %s", err)
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
