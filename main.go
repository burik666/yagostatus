// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/burik666/yagostatus/internal/pkg/config"
)

var builtinConfig = []byte(`
widgets:
  - widget: static
    blocks: >
      [
        {
          "full_text": "YaGoStatus",
          "color": "#2e9ef4"
        }
      ]
    events:
      - button: 1
        command: xdg-open https://github.com/burik666/yagostatus/
  - widget: wrapper
    command: /usr/bin/i3status
  - widget: clock
    format: Jan _2 Mon 15:04:05 # https://golang.org/pkg/time/#Time.Format
    template: >
        {
            "color": "#ffffff",
            "separator": true,
            "separator_block_width": 20
        }
`)

func main() {
	log.SetFlags(log.Ldate + log.Ltime + log.Lshortfile)

	var configFile string
	flag.StringVar(&configFile, "config", "", `config file (default "yagostatus.yml")`)
	versionFlag := flag.Bool("version", false, "print version information and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("YaGoStatus %s\n", Version)
		return
	}

	var cfg *config.Config
	var cfgError, err error

	if configFile == "" {
		cfg, cfgError = config.LoadFile("yagostatus.yml")
		if os.IsNotExist(cfgError) {
			cfgError = nil
			cfg, err = config.Parse(builtinConfig)
			if err != nil {
				log.Fatalf("Failed to parse builtin config: %s", err)
			}
		}
		if cfgError != nil {
			cfg = &config.Config{}
		}
	} else {
		cfg, cfgError = config.LoadFile(configFile)
		if cfgError != nil {
			cfg = &config.Config{}
		}
	}

	yaGoStatus, err := NewYaGoStatus(*cfg)
	if err != nil {
		log.Fatalf("Failed to create yagostatus instance: %s", err)
	}
	if cfgError != nil {
		log.Printf("Failed to load config: %s", cfgError)
		yaGoStatus.errorWidget(cfgError.Error())
	}

	stopContSignals := make(chan os.Signal, 1)
	signal.Notify(stopContSignals, cfg.Signals.StopSignal, cfg.Signals.ContSignal)
	go func() {
		for {
			sig := <-stopContSignals
			switch sig {
			case cfg.Signals.StopSignal:
				yaGoStatus.Stop()
			case cfg.Signals.ContSignal:
				yaGoStatus.Continue()
			}
		}
	}()

	shutdownsignals := make(chan os.Signal, 1)
	signal.Notify(shutdownsignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		yaGoStatus.Run()
		shutdownsignals <- syscall.SIGTERM
	}()

	<-shutdownsignals
	yaGoStatus.Shutdown()
}
