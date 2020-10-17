// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/burik666/yagostatus/internal/pkg/config"
	"github.com/burik666/yagostatus/internal/pkg/logger"
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
    templates: >
        [{
            "color": "#ffffff",
            "separator": true,
            "separator_block_width": 21
        }]
`)

func main() {
	logger := logger.New(log.Ldate + log.Ltime + log.Lshortfile)

	var configFile string

	flag.StringVar(&configFile, "config", "", `config file (default "yagostatus.yml")`)

	versionFlag := flag.Bool("version", false, "print version information and exit")
	swayFlag := flag.Bool("sway", false, "set it when using sway")

	flag.Parse()

	if *versionFlag {
		logger.Infof("YaGoStatus %s", Version)
		return
	}

	var cfg *config.Config

	var cfgError, err error

	if configFile == "" {
		cfg, cfgError = config.LoadFile("yagostatus.yml")
		if os.IsNotExist(cfgError) {
			cfgError = nil

			cfg, err = config.Parse(builtinConfig, "builtin")
			if err != nil {
				logger.Errorf("Failed to parse builtin config: %s", err)
				os.Exit(1)
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

	yaGoStatus, err := NewYaGoStatus(*cfg, *swayFlag, logger)
	if err != nil {
		logger.Errorf("Failed to create yagostatus instance: %s", err)
		os.Exit(1)
	}

	if cfgError != nil {
		logger.Errorf("Failed to load config: %s", cfgError)
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

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := yaGoStatus.Run(); err != nil {
			logger.Errorf("Failed to run yagostatus: %s", err)
		}
		shutdownSignals <- syscall.SIGTERM
	}()

	<-shutdownSignals

	if err := yaGoStatus.Shutdown(); err != nil {
		logger.Errorf("Failed to shutdown yagostatus: %s", err)
	}

	logger.Infof("exit")
}
