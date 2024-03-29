package executor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/burik666/yagostatus/ygs"
)

type OutputFormat string

const (
	OutputFormatAuto OutputFormat = "auto"
	OutputFormatNone OutputFormat = "none"
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
)

type Executor struct {
	cmd    *exec.Cmd
	header *ygs.I3BarHeader

	finished bool
	waiterr  error
}

func Exec(command string, args ...string) (*Executor, error) {
	r := regexp.MustCompile("'.+'|\".+\"|\\S+")
	m := r.FindAllString(command, -1)
	name := m[0]
	args = append(m[1:], args...)

	e := &Executor{}

	e.cmd = exec.Command(name, args...)
	e.cmd.Env = os.Environ()
	e.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	return e, nil
}

func (e *Executor) SetWD(wd string) {
	if e.cmd != nil {
		e.cmd.Dir = wd
	}
}

func (e *Executor) Run(logger ygs.Logger, c chan<- []ygs.I3BarBlock, format OutputFormat) error {
	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return err
	}

	defer stderr.Close()

	go (func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logger.Errorf("(stderr) %s", scanner.Text())
		}
	})()

	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	defer stdout.Close()

	if err := e.cmd.Start(); err != nil {
		return err
	}

	defer func() {
		_ = e.wait()
	}()

	if format == OutputFormatNone {
		return nil
	}

	buf := &bufferCloser{}
	outreader := io.TeeReader(stdout, buf)

	decoder := json.NewDecoder(outreader)

	var firstMessage interface{}

	err = decoder.Decode(&firstMessage)
	if (err != nil) && format == OutputFormatJSON {
		buf.Close()

		return err
	}

	isJSON := false
	switch firstMessage.(type) {
	case map[string]interface{}:
		isJSON = true
	case []interface{}:
		isJSON = true
	}

	if err != nil || !isJSON || format == OutputFormatText {
		_, err := io.Copy(ioutil.Discard, outreader)
		if err != nil {
			return err
		}

		if buf.Len() > 0 {
			lines := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
			out := make([]ygs.I3BarBlock, len(lines))

			for i := range lines {
				out[i] = ygs.I3BarBlock{
					FullText: strings.Trim(lines[i], "\n "),
				}
			}
			c <- out
		}

		buf.Close()

		return nil
	}

	buf.Close()

	firstMessageData, _ := json.Marshal(firstMessage)

	headerDecoder := json.NewDecoder(bytes.NewBuffer(firstMessageData))
	headerDecoder.DisallowUnknownFields()

	var header ygs.I3BarHeader
	if err := headerDecoder.Decode(&header); err == nil {
		e.header = &header

		_, err := decoder.Token()
		if err != nil {
			return err
		}
	} else {
		var blocks []ygs.I3BarBlock
		if err := json.Unmarshal(firstMessageData, &blocks); err != nil {
			return err
		}
		c <- blocks
	}

	defer func() {
		_ = e.Shutdown()
	}()

	for {
		var blocks []ygs.I3BarBlock
		if err := decoder.Decode(&blocks); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			if e.finished {
				return nil
			}

			return err
		}
		c <- blocks
	}
}

func (e *Executor) Stdin() (io.WriteCloser, error) {
	return e.cmd.StdinPipe()
}

func (e *Executor) AddEnv(env ...string) {
	e.cmd.Env = append(e.cmd.Env, env...)
}

func (e *Executor) wait() error {
	if e.finished {
		return e.waiterr
	}

	e.waiterr = e.cmd.Wait()
	e.finished = true

	return e.waiterr
}

func (e *Executor) Shutdown() error {
	if e.finished {
		return nil
	}

	if e.cmd != nil && e.cmd.Process != nil && e.cmd.Process.Pid > 1 {
		return syscall.Kill(-e.cmd.Process.Pid, syscall.SIGTERM)
	}

	return e.wait()
}

func (e *Executor) Signal(sig syscall.Signal) error {
	if e.cmd != nil && e.cmd.Process != nil && e.cmd.Process.Pid > 1 {
		return syscall.Kill(-e.cmd.Process.Pid, sig)
	}

	return nil
}

func (e *Executor) ProcessState() *os.ProcessState {
	return e.cmd.ProcessState
}

func (e *Executor) I3BarHeader() *ygs.I3BarHeader {
	return e.header
}

type bufferCloser struct {
	bytes.Buffer
	stoped bool
}

func (b *bufferCloser) Write(p []byte) (n int, err error) {
	if b.stoped {
		return len(p), nil
	}

	return b.Buffer.Write(p)
}

func (b *bufferCloser) Close() error {
	b.stoped = true
	b.Reset()

	return nil
}
