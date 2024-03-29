package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/burik666/yagostatus/internal/config"
	"github.com/burik666/yagostatus/internal/registry"
	"github.com/burik666/yagostatus/pkg/executor"
	"github.com/burik666/yagostatus/ygs"

	_ "github.com/burik666/yagostatus/plugins"
	_ "github.com/burik666/yagostatus/widgets"

	"go.i3wm.org/i3/v4"
)

type widgetContainer struct {
	instance ygs.Widget
	output   []ygs.I3BarBlock
	config   config.WidgetConfig
	ch       chan []ygs.I3BarBlock
	logger   ygs.Logger
	m        sync.RWMutex
}

// YaGoStatus is the main struct.
type YaGoStatus struct {
	widgets []widgetContainer

	upd chan int

	workspaces        []i3.Workspace
	visibleWorkspaces []string

	cfg  config.Config
	sway bool

	logger ygs.Logger
}

// NewYaGoStatus returns a new YaGoStatus instance.
func NewYaGoStatus(cfg config.Config, sway bool, l ygs.Logger) *YaGoStatus {
	status := &YaGoStatus{
		cfg:    cfg,
		sway:   sway,
		logger: l,
	}

	if sway {
		i3.SocketPathHook = func() (string, error) {
			out, err := exec.Command("sway", "--get-socketpath").CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("getting sway socketpath: %w (output: %s)", err, out)
			}

			return string(out), nil
		}

		i3.IsRunningHook = func() bool {
			out, err := exec.Command("pgrep", "-c", "sway\\$").CombinedOutput()
			if err != nil {
				l.Errorf("sway running: %w (output: %s)", err, out)
			}

			return bytes.Equal(out, []byte("1"))
		}
	}

	for wi := range cfg.Widgets {
		status.addWidget(cfg.Widgets[wi])
	}

	return status
}

func (status *YaGoStatus) errorWidget(text string) {
	status.addWidget(config.ErrorWidget(text))
}

func (status *YaGoStatus) addWidget(wcfg config.WidgetConfig) {
	wlogger := status.logger.WithPrefix(fmt.Sprintf("[%s#%d]", wcfg.File, wcfg.Index+1))

	(func() {
		defer (func() {
			if r := recover(); r != nil {
				wlogger.Errorf("NewWidget panic: %s", r)
				debug.PrintStack()
				status.errorWidget("widget panic")
			}
		})()

		widget, err := registry.NewWidget(wcfg, wlogger)
		if err != nil {
			wlogger.Errorf("Failed to create widget: %s", err)
			status.errorWidget(err.Error())

			return
		}

		status.widgets = append(status.widgets, widgetContainer{
			instance: widget,
			config:   wcfg,
			logger:   wlogger,
		})
	})()
}

func (status *YaGoStatus) processWidgetEvents(wi int, block ygs.I3BarBlock, event ygs.I3BarClickEvent) error {
	defer (func() {
		if r := recover(); r != nil {
			status.widgets[wi].logger.Errorf("widget event panic: %s", r)
			debug.PrintStack()

			status.widgets[wi].ch <- []ygs.I3BarBlock{{
				FullText: "widget panic",
				Color:    "#ff0000",
			}}
		}

		status.widgets[wi].m.Lock()

		blocks := make([]ygs.I3BarBlock, len(status.widgets[wi].output))
		copy(blocks, status.widgets[wi].output)

		status.widgets[wi].m.Unlock()

		if err := status.widgets[wi].instance.Event(event, blocks); err != nil {
			status.widgets[wi].logger.Errorf("Failed to process widget event: %s", err)
		}
	})()

	for _, widgetEvent := range status.widgets[wi].config.Events {
		if (widgetEvent.Button == 0 || widgetEvent.Button == event.Button) &&
			(widgetEvent.Name == "" || widgetEvent.Name == event.Name) &&
			(widgetEvent.Instance == "" || widgetEvent.Instance == event.Instance) &&
			checkModifiers(widgetEvent.Modifiers, event.Modifiers) {
			exc, err := executor.Exec("sh", "-c", widgetEvent.Command)
			if err != nil {
				return err
			}

			exc.SetWD(widgetEvent.WorkDir)

			exc.AddEnv(
				fmt.Sprintf("I3_%s=%s", "NAME", event.Name),
				fmt.Sprintf("I3_%s=%s", "INSTANCE", event.Instance),
				fmt.Sprintf("I3_%s=%d", "BUTTON", event.Button),
				fmt.Sprintf("I3_%s=%d", "X", event.X),
				fmt.Sprintf("I3_%s=%d", "Y", event.Y),
				fmt.Sprintf("I3_%s=%d", "RELATIVE_X", event.RelativeX),
				fmt.Sprintf("I3_%s=%d", "RELATIVE_Y", event.RelativeY),
				fmt.Sprintf("I3_%s=%d", "OUTPUT_X", event.OutputX),
				fmt.Sprintf("I3_%s=%d", "OUTPUT_Y", event.OutputY),
				fmt.Sprintf("I3_%s=%d", "WIDTH", event.Width),
				fmt.Sprintf("I3_%s=%d", "HEIGHT", event.Height),
				fmt.Sprintf("I3_%s=%s", "MODIFIERS", strings.Join(event.Modifiers, ",")),
			)

			exc.AddEnv(widgetEvent.Env...)

			exc.AddEnv(block.Env("")...)

			stdin, err := exc.Stdin()
			if err != nil {
				return err
			}

			eventJSON, err := json.Marshal(event)
			if err != nil {
				return err
			}

			eventJSON = append(eventJSON, []byte("\n")...)

			_, err = stdin.Write(eventJSON)
			if err != nil {
				return err
			}

			err = stdin.Close()
			if err != nil {
				return err
			}

			err = exc.Run(
				status.widgets[wi].logger,
				status.widgets[wi].ch,
				executor.OutputFormat(widgetEvent.OutputFormat),
			)
			if err != nil {
				return err
			}

			if state := exc.ProcessState(); state != nil && state.ExitCode() != 0 {
				return fmt.Errorf("process exited unexpectedly: %s", state.String())
			}
		}
	}

	return nil
}

func (status *YaGoStatus) addWidgetOutput(wi int, blocks []ygs.I3BarBlock) {
	output := make([]ygs.I3BarBlock, len(blocks))
	tplc := len(status.widgets[wi].config.Templates)

	for blockIndex := range blocks {
		block := blocks[blockIndex]

		if tplc == 1 {
			block.Apply(status.widgets[wi].config.Templates[0])
		} else if blockIndex < tplc {
			block.Apply(status.widgets[wi].config.Templates[blockIndex])
		}

		block.Name = fmt.Sprintf("yagostatus-%d-%s", wi, block.Name)
		block.Instance = fmt.Sprintf("yagostatus-%d-%d-%s", wi, blockIndex, block.Instance)

		output[blockIndex] = block
	}

	status.widgets[wi].m.Lock()
	status.widgets[wi].output = output
	status.widgets[wi].m.Unlock()

	status.upd <- wi
}

func (status *YaGoStatus) eventReader() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}

			break
		}

		line = strings.Trim(line, "[], \n")
		if line == "" {
			continue
		}

		var event ygs.I3BarClickEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			status.logger.Errorf("%s (%s)", err, line)

			continue
		}

		go func(event ygs.I3BarClickEvent) {
			wi, name, err := splitName(event.Name)
			if err != nil {
				status.logger.Errorf("failed to parse event name '%s': %s", event.Name, err)

				return
			}

			_, oi, instance, err := splitInstance(event.Instance)
			if err != nil {
				status.logger.Errorf("failed to parse event instance '%s': %s", event.Name, err)

				return
			}

			e := event
			e.Name = name
			e.Instance = instance

			status.widgets[wi].m.RLock()
			if len(status.widgets[wi].output) < oi {
				status.widgets[wi].m.RUnlock()

				return
			}

			block := status.widgets[wi].output[oi]
			status.widgets[wi].m.RUnlock()

			if (event.Name != "" && event.Name == block.Name) && (event.Instance != "" && event.Instance == block.Instance) {
				block.Name = e.Name
				block.Instance = e.Instance

				if err := status.processWidgetEvents(wi, block, e); err != nil {
					status.widgets[wi].logger.Errorf("event error: %s", err)

					status.widgets[wi].ch <- []ygs.I3BarBlock{{
						FullText: fmt.Sprintf("event error: %s", err.Error()),
						Color:    "#ff0000",
						Name:     event.Name,
						Instance: event.Instance,
					}}
				}
			}
		}(event)
	}

	return nil
}

// Run starts the main loop.
func (status *YaGoStatus) Run() error {
	status.upd = make(chan int)

	go (func() {
		status.updateWorkspaces()

		recv := i3.Subscribe(i3.WorkspaceEventType)
		for recv.Next() {
			e := recv.Event().(*i3.WorkspaceEvent)
			if e.Change == "empty" {
				continue
			}

			status.updateWorkspaces()
			status.upd <- -1
		}
	})()

	widgetChans := make([]reflect.SelectCase, len(status.widgets))

	var wcounter int32 = int32(len(status.widgets))

	for wi := range status.widgets {
		status.widgets[wi].ch = make(chan []ygs.I3BarBlock)

		widgetChans[wi] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(status.widgets[wi].ch),
		}

		go func(wi int) {
			defer (func() {
				atomic.AddInt32(&wcounter, -1)
				if r := recover(); r != nil {
					status.widgets[wi].logger.Errorf("widget panic: %s", r)
					debug.PrintStack()
					status.widgets[wi].ch <- []ygs.I3BarBlock{{
						FullText: "widget panic",
						Color:    "#ff0000",
					}}
				}
			})()

			if err := status.widgets[wi].instance.Run(status.widgets[wi].ch); err != nil {
				status.widgets[wi].logger.Errorf("Widget done: %s", err)
				status.widgets[wi].ch <- []ygs.I3BarBlock{{
					FullText: err.Error(),
					Color:    "#ff0000",
				}}
			}
		}(wi)
	}

	go func() {
		for wcounter > 0 {
			wi, out, _ := reflect.Select(widgetChans)
			status.addWidgetOutput(wi, out.Interface().([]ygs.I3BarBlock))
		}
	}()

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(ygs.I3BarHeader{
		Version:     1,
		ClickEvents: true,
		StopSignal:  int(status.cfg.Signals.StopSignal),
		ContSignal:  int(status.cfg.Signals.ContSignal),
	}); err != nil {
		status.logger.Errorf("Failed to encode I3BarHeader: %s", err)
	}

	fmt.Print("\n[\n[]")

	go func() {
		for range status.upd {
			var result []ygs.I3BarBlock

			for wi := range status.widgets {
				if checkWorkspaceConditions(status.widgets[wi].config.Workspaces, status.visibleWorkspaces) {
					status.widgets[wi].m.RLock()
					result = append(result, status.widgets[wi].output...)
					status.widgets[wi].m.RUnlock()
				}
			}

			fmt.Print(",")

			if result == nil {
				fmt.Print("[]")

				continue
			}

			if err := encoder.Encode(result); err != nil {
				status.logger.Errorf("Failed to encode result: %s", err)

				break
			}
		}
	}()

	return status.eventReader()
}

// Shutdown shutdowns widgets and main loop.
func (status *YaGoStatus) Shutdown() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	for wi := range status.widgets {
		wg.Add(1)

		done := make(chan struct{})
		defer close(done)

		go func(wi int) {
			defer wg.Done()

			defer (func() {
				if r := recover(); r != nil {
					status.widgets[wi].logger.Errorf("widget panic: %s", r)
					debug.PrintStack()
				}
			})()

			go func() {
				if err := status.widgets[wi].instance.Shutdown(); err != nil {
					status.widgets[wi].logger.Errorf("Failed to shutdown widget: %s", err)
				}

				done <- struct{}{}
			}()

			select {
			case <-ctx.Done():
				status.widgets[wi].logger.Errorf("Failed to shutdown widget: %s", ctx.Err())
			case <-done:
			}
		}(wi)
	}

	wg.Wait()
}

// Stop stops widgets and main loop.
func (status *YaGoStatus) Stop() {
	for wi := range status.widgets {
		go func(wi int) {
			defer (func() {
				if r := recover(); r != nil {
					status.widgets[wi].logger.Errorf("widget panic: %s", r)
					debug.PrintStack()
				}
			})()

			if err := status.widgets[wi].instance.Stop(); err != nil {
				status.widgets[wi].logger.Errorf("Failed to stop widget: %s", err)
			}
		}(wi)
	}
}

// Continue continues widgets and main loop.
func (status *YaGoStatus) Continue() {
	for wi := range status.widgets {
		go func(wi int) {
			defer (func() {
				if r := recover(); r != nil {
					status.widgets[wi].logger.Errorf("widget panic: %s", r)
					debug.PrintStack()
				}
			})()

			if err := status.widgets[wi].instance.Continue(); err != nil {
				status.widgets[wi].logger.Errorf("Failed to continue widget: %s", err)
			}
		}(wi)
	}
}

func (status *YaGoStatus) updateWorkspaces() {
	var err error

	status.workspaces, err = i3.GetWorkspaces()

	if err != nil {
		status.logger.Errorf("Failed to get workspaces: %s", err)
	}

	var vw []string

	for i := range status.workspaces {
		if status.workspaces[i].Visible {
			vw = append(vw, status.workspaces[i].Name)
		}
	}

	status.visibleWorkspaces = vw
}

func checkModifiers(conditions []string, values []string) bool {
	for _, c := range conditions {
		isNegative := c[0] == '!'
		c = strings.TrimLeft(c, "!")

		found := false

		for _, v := range values {
			if c == v {
				found = true

				break
			}
		}

		if found && isNegative {
			return false
		}

		if !found && !isNegative {
			return false
		}
	}

	return true
}

func checkWorkspaceConditions(conditions []string, values []string) bool {
	if len(conditions) == 0 {
		return true
	}

	pass := 0

	for _, c := range conditions {
		isNegative := c[0] == '!'
		c = strings.TrimLeft(c, "!")

		found := false

		for _, v := range values {
			if c == v {
				found = true

				break
			}
		}

		if found && !isNegative {
			return true
		}

		if !found && isNegative {
			pass++
		}
	}

	return len(conditions) == pass
}

func splitName(name string) (int, string, error) {
	parts := strings.SplitN(name, "-", 3)

	wi, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, "", err
	}

	return int(wi), parts[2], nil
}

func splitInstance(name string) (int, int, string, error) {
	parts := strings.SplitN(name, "-", 4)

	wi, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, "", err
	}

	oi, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, "", err
	}

	return int(wi), int(oi), parts[3], nil
}
