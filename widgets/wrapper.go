package widgets

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/burik666/yagostatus/ygs"
	"io"
	"os"
	"os/exec"
	"regexp"
)

type WrapperWidget struct {
	stdin   io.WriteCloser
	cmd     *exec.Cmd
	command string
	args    []string
}

func (w *WrapperWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["command"]
	if !ok {
		return errors.New("Missing 'command' setting")
	}
	r := regexp.MustCompile("'.+'|\".+\"|\\S+")
	m := r.FindAllString(v.(string), -1)
	w.command = m[0]
	w.args = m[1:]

	return nil
}

func (w *WrapperWidget) Run(c chan []ygs.I3BarBlock) error {
	w.cmd = exec.Command(w.command, w.args...)
	w.cmd.Stderr = os.Stderr
	stdout, err := w.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := w.cmd.Start(); err != nil {
		return err
	}
	w.stdin.Write([]byte("["))

	reader := bufio.NewReader(stdout)
	decoder := json.NewDecoder(reader)

	var header ygs.I3BarHeader
	if err := decoder.Decode(&header); err == nil {
		decoder.Token()
	}

	for {
		var blocks []ygs.I3BarBlock
		err := decoder.Decode(&blocks)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		c <- blocks
	}
	return w.cmd.Wait()
}

func (w *WrapperWidget) Event(event ygs.I3BarClickEvent) {
	msg, _ := json.Marshal(event)
	w.stdin.Write(msg)
	w.stdin.Write([]byte(",\n"))
}

func (w *WrapperWidget) Stop() {
}

func init() {
	ygs.RegisterWidget(WrapperWidget{})
}
