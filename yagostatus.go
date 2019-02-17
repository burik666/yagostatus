package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/burik666/yagostatus/ygs"
)

// A YaGoStatus is the main struct.
type YaGoStatus struct {
	widgets       []ygs.Widget
	widgetsOutput [][]ygs.I3BarBlock
	widgetsConfig []ConfigWidget
	upd           chan int
}

// Configure loads config from a file.
func (status *YaGoStatus) Configure(configFile string) error {
	config, err := loadConfig(configFile)
	if err != nil {
		status.errorWidget(err.Error())
		config, err = parseConfig([]byte(`
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
        }`))
		if err != nil {
			return err
		}
	}
	for _, w := range config.Widgets {
		widget, ok := ygs.NewWidget(w.Name + "widget")
		if !ok {
			status.errorWidget(fmt.Sprintf("Widget '%s' not found", w.Name))
			continue
		}

		err := status.AddWidget(widget, w)
		if err != nil {
			status.errorWidget(fmt.Sprintf("Widget '%s' configuration error: %s", w.Name, err))
			continue
		}
	}
	return nil
}

func (status *YaGoStatus) errorWidget(text string) {
	log.Print(text)
	widget, ok := ygs.NewWidget("staticwidget")
	if !ok {
		log.Fatal("Failed to create error widget: 'staticwidget' not found")
	}
	blocks, _ := json.Marshal([]ygs.I3BarBlock{
		ygs.I3BarBlock{
			FullText: text,
			Color:    "#ff0000",
		},
	})
	err := status.AddWidget(widget, ConfigWidget{
		Params: map[string]interface{}{
			"blocks": string(blocks),
		},
	})
	if err != nil {
		log.Fatalf("Failed to configure error widget: %s", err)
	}
}

// AddWidget adds widget to statusbar.
func (status *YaGoStatus) AddWidget(widget ygs.Widget, config ConfigWidget) error {
	if err := widget.Configure(config.Params); err != nil {
		return err
	}
	status.widgets = append(status.widgets, widget)
	status.widgetsOutput = append(status.widgetsOutput, nil)
	status.widgetsConfig = append(status.widgetsConfig, config)

	return nil
}

func (status *YaGoStatus) processWidgetEvents(widgetIndex int, outputIndex int, event ygs.I3BarClickEvent) error {
	defer status.widgets[widgetIndex].Event(event)
	for _, we := range status.widgetsConfig[widgetIndex].Events {
		if (we.Button == 0 || we.Button == event.Button) &&
			(we.Name == "" || we.Name == event.Name) &&
			(we.Instance == "" || we.Instance == event.Instance) &&
			checkModifiers(we.Modifiers, event.Modifiers) {
			cmd := exec.Command("sh", "-c", we.Command)
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(),
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
			cmdStdin, err := cmd.StdinPipe()
			if err != nil {
				return err
			}
			eventJSON, _ := json.Marshal(event)
			cmdStdin.Write(eventJSON)
			cmdStdin.Write([]byte("\n"))
			cmdStdin.Close()

			cmdOutput, err := cmd.Output()
			if err != nil {
				return err
			}
			if we.Output {
				var blocks []ygs.I3BarBlock
				if err := json.Unmarshal(cmdOutput, &blocks); err == nil {
					for bi := range blocks {
						block := &blocks[bi]
						mergeBlocks(block, status.widgetsConfig[widgetIndex].Template)
						block.Name = fmt.Sprintf("ygs-%d-%s", widgetIndex, block.Name)
						block.Instance = fmt.Sprintf("ygs-%d-%d-%s", widgetIndex, outputIndex, block.Instance)
					}
					status.widgetsOutput[widgetIndex] = blocks
				} else {
					status.widgetsOutput[widgetIndex][outputIndex].FullText = strings.Trim(string(cmdOutput), "\n\r")
				}
				status.upd <- widgetIndex
			}
		}
	}
	return nil
}

func (status *YaGoStatus) eventReader() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		line = strings.Trim(line, "[], \n")
		if line != "" {
			var event ygs.I3BarClickEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				log.Printf("%s (%s)", err, line)
			} else {
				for i, widgetOutputs := range status.widgetsOutput {
					for j, output := range widgetOutputs {
						if (event.Name != "" && event.Name == output.Name) && (event.Instance != "" && event.Instance == output.Instance) {
							e := event
							e.Name = strings.Join(strings.Split(e.Name, "-")[2:], "-")
							e.Instance = strings.Join(strings.Split(e.Instance, "-")[3:], "-")
							if err := status.processWidgetEvents(i, j, e); err != nil {
								log.Print(err)
								status.widgetsOutput[i][j] = ygs.I3BarBlock{
									FullText: fmt.Sprintf("Event error: %s", err.Error()),
									Color:    "#ff0000",
									Name:     event.Name,
									Instance: event.Instance,
								}
								break

							}
						}
					}
				}
			}
		}
	}
}

// Run starts the main loop.
func (status *YaGoStatus) Run() {
	status.upd = make(chan int)
	for i, widget := range status.widgets {
		c := make(chan []ygs.I3BarBlock)
		go func(i int, c chan []ygs.I3BarBlock) {
			for {
				select {
				case out := <-c:
					output := make([]ygs.I3BarBlock, len(out))
					copy(output, out)
					for j := range output {
						mergeBlocks(&output[j], status.widgetsConfig[i].Template)
						output[j].Name = fmt.Sprintf("ygs-%d-%s", i, output[j].Name)
						output[j].Instance = fmt.Sprintf("ygs-%d-%d-%s", i, j, output[j].Instance)
					}
					status.widgetsOutput[i] = output
					status.upd <- i
				}
			}
		}(i, c)

		go func(widget ygs.Widget, c chan []ygs.I3BarBlock) {
			if err := widget.Run(c); err != nil {
				log.Print(err)
				c <- []ygs.I3BarBlock{ygs.I3BarBlock{
					FullText: err.Error(),
					Color:    "#ff0000",
				}}
			}
		}(widget, c)
	}

	fmt.Print("{\"version\":1, \"click_events\": true}\n[\n[]")
	go func() {
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		for {
			select {
			case <-status.upd:
				var result []ygs.I3BarBlock
				for _, widgetOutput := range status.widgetsOutput {
					result = append(result, widgetOutput...)
				}
				buf.Reset()
				encoder.Encode(result)
				fmt.Print(",")
				fmt.Print(string(buf.Bytes()))
			}
		}
	}()
	status.eventReader()
}

// Stop shutdowns widgets and main loop.
func (status *YaGoStatus) Stop() {
	var wg sync.WaitGroup
	for _, widget := range status.widgets {
		wg.Add(1)
		go func(widget ygs.Widget) {
			widget.Stop()
			wg.Done()
		}(widget)
	}
	wg.Wait()
}

func mergeBlocks(b *ygs.I3BarBlock, tpl ygs.I3BarBlock) {
	var resmap map[string]interface{}

	jb, _ := json.Marshal(*b)
	jtpl, _ := json.Marshal(tpl)
	json.Unmarshal(jtpl, &resmap)
	json.Unmarshal(jb, &resmap)

	jb, _ = json.Marshal(resmap)
	json.Unmarshal(jb, b)
}

func checkModifiers(conditions []string, modifiers []string) bool {
	for _, c := range conditions {
		isNegative := c[0] == '!'
		c = strings.TrimLeft(c, "!")
		found := false
		for _, v := range modifiers {
			if c == v {
				found = true
				break
			}
		}
		if found && isNegative {
			return false
		}
		if (!found) && !isNegative {
			return false
		}
	}
	return true
}
