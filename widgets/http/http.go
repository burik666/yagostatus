package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denysvitali/yagostatus/widgets/blank"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/denysvitali/yagostatus/internal/pkg/logger"
	"github.com/denysvitali/yagostatus/ygs"

	"golang.org/x/net/websocket"
)

// HTTPWidgetParams are widget parameters.
type HTTPWidgetParams struct {
	Network string
	Listen  string
	Path    string
}

var _ ygs.Widget = &HTTPWidget{}

// HTTPWidget implements the http server widget.
type HTTPWidget struct {
	blank.Widget

	params HTTPWidgetParams

	logger logger.Logger

	c        chan<- []ygs.I3BarBlock
	instance *httpInstance

	clients map[*websocket.Conn]chan interface{}
	cm      sync.RWMutex
}

type httpInstance struct {
	l      net.Listener
	server *http.Server
	mux    *http.ServeMux
	paths  map[string]struct{}
}

var instances map[string]*httpInstance

func init() {
	ygs.RegisterWidget("http", NewHTTPWidget, HTTPWidgetParams{
		Network: "tcp",
	})

	instances = make(map[string]*httpInstance, 1)
}

// NewHTTPWidget returns a new HTTPWidget.
func NewHTTPWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	w := &HTTPWidget{
		params: params.(HTTPWidgetParams),
		logger: wlogger,
	}

	if len(w.params.Listen) == 0 {
		return nil, errors.New("missing 'listen'")
	}

	if len(w.params.Path) == 0 {
		return nil, errors.New("missing 'path'")
	}

	if w.params.Network != "tcp" && w.params.Network != "unix" {
		return nil, errors.New("invalid 'net' (may be 'tcp' or 'unix')")
	}

	instanceKey := w.params.Listen
	instance, ok := instances[instanceKey]
	if ok {
		if _, ok := instance.paths[w.params.Path]; ok {
			return nil, fmt.Errorf("path '%s' already in use", w.params.Path)
		}
	} else {
		mux := http.NewServeMux()
		instance = &httpInstance{
			mux:   mux,
			paths: make(map[string]struct{}, 1),
			server: &http.Server{
				Addr:    w.params.Listen,
				Handler: mux,
			},
		}

		instances[w.params.Listen] = instance
		w.instance = instance
	}

	instance.mux.HandleFunc(w.params.Path, w.httpHandler)
	instance.paths[instanceKey] = struct{}{}

	w.clients = make(map[*websocket.Conn]chan interface{})
	return w, nil
}

// Run starts the main loop.
func (w *HTTPWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c

	if w.instance == nil {
		return nil
	}

	l, err := net.Listen(w.params.Network, w.params.Listen)
	if err != nil {
		return err
	}

	w.instance.l = l

	err = w.instance.server.Serve(l)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

// Event processes the widget events.
func (w *HTTPWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	return w.broadcast(event)
}

func (w *HTTPWidget) Shutdown() error {
	if w.instance == nil {
		return nil
	}

	return w.instance.server.Shutdown(context.Background())
}

func (w *HTTPWidget) httpHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		serv := websocket.Server{
			Handshake: func(cfg *websocket.Config, r *http.Request) error {
				return nil
			},
			Handler: w.wsHandler,
		}

		serv.ServeHTTP(response, request)

		return
	}

	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			w.logger.Errorf("%s", err)
		}

		var messages []ygs.I3BarBlock
		if err := json.Unmarshal(body, &messages); err != nil {
			w.logger.Errorf("%s", err)
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, "%s", err)
		}

		w.c <- messages

		return
	}

	response.WriteHeader(http.StatusBadRequest)

	_, err := response.Write([]byte("bad request method, allow GET for websocket and POST for HTTP update"))
	if err != nil {
		w.logger.Errorf("failed to write response: %s", err)
	}
}

func (w *HTTPWidget) wsHandler(ws *websocket.Conn) {
	defer ws.Close()

	ch := make(chan interface{})

	w.cm.RLock()
	w.clients[ws] = ch
	w.cm.RUnlock()

	var blocks []ygs.I3BarBlock

	go func() {
		for {
			msg, ok := <-ch
			if !ok {
				return
			}

			if err := websocket.JSON.Send(ws, msg); err != nil {
				w.logger.Errorf("failed to send msg: %s", err)
			}
		}
	}()

	for {
		if err := websocket.JSON.Receive(ws, &blocks); err != nil {
			if err == io.EOF {
				break
			}

			w.logger.Errorf("invalid message: %s", err)
			break
		}

		w.c <- blocks
	}

	w.cm.Lock()
	delete(w.clients, ws)
	w.cm.Unlock()

	close(ch)
}

func (w *HTTPWidget) broadcast(msg interface{}) error {
	w.cm.RLock()
	defer w.cm.RUnlock()

	for _, ch := range w.clients {
		ch <- msg
	}

	return nil
}
