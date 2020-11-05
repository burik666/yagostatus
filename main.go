// Yet Another i3status replacement written in Go.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/burik666/yagostatus/internal/config"
	"github.com/burik666/yagostatus/internal/logger"
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
	logger := logger.New()

	var configFile string

	flag.StringVar(&configFile, "config", "", `config file (default "yagostatus.yml")`)

	versionFlag := flag.Bool("version", false, "print version information and exit")
	swayFlag := flag.Bool("sway", false, "set it when using sway")
	dumpConfigFlag := flag.Bool("dump", false, "dump parsed config file to stdout")

	flag.Parse()

	if *versionFlag {
		logger.Infof("YaGoStatus %s", Version)

		return
	}

	var initErrors []error

	cfg, cfgError := loadConfig(configFile)
	if cfgError != nil {
		logger.Errorf("Failed to load config: %s", cfgError)
		initErrors = append(initErrors, cfgError)
	}

	if cfg != nil {
		logger.Infof("using config: %s", cfg.File)
	} else {
		cfg = &config.Config{}
	}

	if err := config.LoadPlugins(*cfg, logger); err != nil {
		logger.Errorf("Failed to load plugins: %s", err)
		initErrors = append(initErrors, err)
	}

	if *dumpConfigFlag {
		b, err := config.Dump(cfg)
		if err != nil {
			logger.Errorf("Failed to dump config: %s", err)
			os.Exit(1)
		}

		_, _ = os.Stdout.Write(b)
		os.Exit(0)
	}

	if err := config.InitPlugins(logger); err != nil {
		logger.Errorf("Failed to init plugins: %s", err)
		initErrors = append(initErrors, err)
	}

	yaGoStatus := NewYaGoStatus(*cfg, *swayFlag, logger)

	for _, err := range initErrors {
		yaGoStatus.errorWidget(err.Error())
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
	signal.Notify(shutdownsignals,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGPIPE,
	)

	go func() {
		if err := yaGoStatus.Run(); err != nil {
			logger.Errorf("Failed to run yagostatus: %s", err)
		}
		shutdownsignals <- syscall.SIGTERM
	}()

	<-shutdownsignals

	logger.Infof("shutdown")

	config.ShutdownPlugins(logger)

	yaGoStatus.Shutdown()

	logger.Infof("exit")
}

func loadConfig(configFile string) (*config.Config, error) {
	if configFile == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config dir: %s", err)
		}

		cfg, err := config.LoadFile(configDir + "/yagostatus/yagostatus.yml")
		if os.IsNotExist(err) {
			cfg, err := config.LoadFile("yagostatus.yml")
			if os.IsNotExist(err) {
				return config.Parse(builtinConfig, "builtin")
			}

			return cfg, err
		}

		return cfg, err
	}

	return config.LoadFile(configFile)
}
