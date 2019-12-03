package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/burik666/yagostatus/internal/pkg/config"
	"github.com/burik666/yagostatus/internal/pkg/executor"
	_ "github.com/burik666/yagostatus/widgets"
	"github.com/burik666/yagostatus/ygs"

	"go.i3wm.org/i3"
)

// YaGoStatus is the main struct.
type YaGoStatus struct {
	widgets       []ygs.Widget
	widgetsOutput [][]ygs.I3BarBlock
	widgetsConfig []ygs.WidgetConfig
	widgetChans   []chan []ygs.I3BarBlock

	upd chan int

	workspaces        []i3.Workspace
	visibleWorkspaces []string

	cfg config.Config
}

// NewYaGoStatus returns a new YaGoStatus instance.
func NewYaGoStatus(cfg config.Config) (*YaGoStatus, error) {
	status := &YaGoStatus{cfg: cfg}

	for _, w := range cfg.Widgets {
		(func() {
			defer (func() {
				if r := recover(); r != nil {
					log.Printf("NewWidget is panicking: %s", r)
					debug.PrintStack()
					status.errorWidget("Widget is panicking")
				}
			})()

			widget, err := ygs.NewWidget(w)
			if err != nil {
				log.Printf("Failed to create widget: %s", err)
				status.errorWidget(err.Error())

				return
			}

			status.AddWidget(widget, w)
		})()
	}

	return status, nil
}

func (status *YaGoStatus) errorWidget(text string) {
	errWidget, err := ygs.NewWidget(ygs.ErrorWidget(text))
	if err != nil {
		panic(err)
	}

	status.AddWidget(errWidget, ygs.WidgetConfig{})
}

// AddWidget adds widget to statusbar.
func (status *YaGoStatus) AddWidget(widget ygs.Widget, config ygs.WidgetConfig) {
	status.widgets = append(status.widgets, widget)
	status.widgetsOutput = append(status.widgetsOutput, nil)
	status.widgetsConfig = append(status.widgetsConfig, config)
}

func (status *YaGoStatus) processWidgetEvents(widgetIndex int, outputIndex int, event ygs.I3BarClickEvent) error {
	defer (func() {
		if r := recover(); r != nil {
			log.Printf("Widget event is panicking: %s", r)
			debug.PrintStack()
			status.widgetsOutput[widgetIndex] = []ygs.I3BarBlock{{
				FullText: "Widget event is panicking",
				Color:    "#ff0000",
			}}
		}

		if err := status.widgets[widgetIndex].Event(event, status.widgetsOutput[widgetIndex]); err != nil {
			log.Printf("Failed to process widget event: %s", err)
		}
	})()

	for _, widgetEvent := range status.widgetsConfig[widgetIndex].Events {
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
				fmt.Sprintf("I3_%s=%d", "WIDTH", event.Width),
				fmt.Sprintf("I3_%s=%d", "HEIGHT", event.Height),
				fmt.Sprintf("I3_%s=%s", "MODIFIERS", strings.Join(event.Modifiers, ",")),
			)

			for k, v := range status.widgetsOutput[widgetIndex][outputIndex].Custom {
				exc.AddEnv(
					fmt.Sprintf("I3_%s=%s", k, v),
				)
			}

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

			err = exc.Run(status.widgetChans[widgetIndex], executor.OutputFormat(widgetEvent.OutputFormat))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (status *YaGoStatus) addWidgetOutput(widgetIndex int, blocks []ygs.I3BarBlock) {
	output := make([]ygs.I3BarBlock, len(blocks))

	tplc := len(status.widgetsConfig[widgetIndex].Templates)
	for blockIndex := range blocks {
		block := blocks[blockIndex]

		if tplc == 1 {
			block.Apply(status.widgetsConfig[widgetIndex].Templates[0])
		} else {
			if blockIndex < tplc {
				block.Apply(status.widgetsConfig[widgetIndex].Templates[blockIndex])
			}
		}

		block.Name = fmt.Sprintf("yagostatus-%d-%s", widgetIndex, block.Name)
		block.Instance = fmt.Sprintf("yagostatus-%d-%d-%s", widgetIndex, blockIndex, block.Instance)

		output[blockIndex] = block
	}

	status.widgetsOutput[widgetIndex] = output

	status.upd <- widgetIndex
}

func (status *YaGoStatus) eventReader() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
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
			log.Printf("%s (%s)", err, line)

			continue
		}

		for widgetIndex, widgetOutputs := range status.widgetsOutput {
			for outputIndex, output := range widgetOutputs {
				if (event.Name != "" && event.Name == output.Name) && (event.Instance != "" && event.Instance == output.Instance) {
					e := event
					e.Name = strings.Join(strings.Split(e.Name, "-")[2:], "-")
					e.Instance = strings.Join(strings.Split(e.Instance, "-")[3:], "-")

					if err := status.processWidgetEvents(widgetIndex, outputIndex, e); err != nil {
						log.Print(err)

						status.widgetsOutput[widgetIndex][outputIndex] = ygs.I3BarBlock{
							FullText: fmt.Sprintf("Event error: %s", err.Error()),
							Color:    "#ff0000",
							Name:     event.Name,
							Instance: event.Instance,
						}
					}

					break
				}
			}
		}
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

	for widgetIndex, widget := range status.widgets {
		c := make(chan []ygs.I3BarBlock)
		status.widgetChans = append(status.widgetChans, c)

		go func(widgetIndex int, c chan []ygs.I3BarBlock) {
			for out := range c {
				status.addWidgetOutput(widgetIndex, out)
			}
		}(widgetIndex, c)

		go func(widget ygs.Widget, c chan []ygs.I3BarBlock) {
			defer (func() {
				if r := recover(); r != nil {
					c <- []ygs.I3BarBlock{{
						FullText: "Widget is panicking",
						Color:    "#ff0000",
					}}
					log.Printf("Widget is panicking: %s", r)
					debug.PrintStack()
				}
			})()
			if err := widget.Run(c); err != nil {
				log.Print(err)
				c <- []ygs.I3BarBlock{{
					FullText: err.Error(),
					Color:    "#ff0000",
				}}
			}
		}(widget, c)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(ygs.I3BarHeader{
		Version:     1,
		ClickEvents: true,
		StopSignal:  int(status.cfg.Signals.StopSignal),
		ContSignal:  int(status.cfg.Signals.ContSignal),
	}); err != nil {
		log.Printf("Failed to encode I3BarHeader: %s", err)
	}

	fmt.Print("\n[\n[]")

	go func() {
		for range status.upd {
			var result []ygs.I3BarBlock
			for widgetIndex, widgetOutput := range status.widgetsOutput {
				if checkWorkspaceConditions(status.widgetsConfig[widgetIndex].Workspaces, status.visibleWorkspaces) {
					result = append(result, widgetOutput...)
				}
			}
			fmt.Print(",")
			err := encoder.Encode(result)
			if err != nil {
				log.Printf("Failed to encode result: %s", err)
			}
		}
	}()

	return status.eventReader()
}

// Shutdown shutdowns widgets and main loop.
func (status *YaGoStatus) Shutdown() error {
	var wg sync.WaitGroup

	for _, widget := range status.widgets {
		wg.Add(1)

		go func(widget ygs.Widget) {
			defer wg.Done()
			defer (func() {
				if r := recover(); r != nil {
					log.Printf("Widget is panicking: %s", r)
					debug.PrintStack()
				}
			})()
			if err := widget.Shutdown(); err != nil {
				log.Printf("Failed to shutdown widget: %s", err)
			}
		}(widget)
	}

	wg.Wait()

	return nil
}

// Stop stops widgets and main loop.
func (status *YaGoStatus) Stop() {
	for _, widget := range status.widgets {
		go func(widget ygs.Widget) {
			defer (func() {
				if r := recover(); r != nil {
					log.Printf("Widget is panicking: %s", r)
					debug.PrintStack()
				}
			})()
			if err := widget.Stop(); err != nil {
				log.Printf("Failed to stop widget: %s", err)
			}
		}(widget)
	}
}

// Continue continues widgets and main loop.
func (status *YaGoStatus) Continue() {
	for _, widget := range status.widgets {
		go func(widget ygs.Widget) {
			defer (func() {
				if r := recover(); r != nil {
					log.Printf("Widget is panicking: %s", r)
					debug.PrintStack()
				}
			})()
			if err := widget.Continue(); err != nil {
				log.Printf("Failed to continue widget: %s", err)
			}
		}(widget)
	}
}

func (status *YaGoStatus) updateWorkspaces() {
	var err error

	status.workspaces, err = i3.GetWorkspaces()

	if err != nil {
		log.Printf("Failed to get workspaces: %s", err)
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
