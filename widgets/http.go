package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/burik666/yagostatus/ygs"

	"golang.org/x/net/websocket"
)

// HTTPWidgetParams are widget parameters.
type HTTPWidgetParams struct {
	Listen string
	Path   string
}

// HTTPWidget implements the http server widget.
type HTTPWidget struct {
	BlankWidget

	params HTTPWidgetParams

	httpServer *http.Server
	conn       *websocket.Conn
	c          chan<- []ygs.I3BarBlock
}

type httpInstance struct {
	mux   *http.ServeMux
	paths map[string]struct{}
}

var instances map[string]*httpInstance

func init() {
	ygs.RegisterWidget("http", NewHTTPWidget, HTTPWidgetParams{})
}

// NewHTTPWidget returns a new HTTPWidget.
func NewHTTPWidget(params interface{}) (ygs.Widget, error) {
	w := &HTTPWidget{
		params: params.(HTTPWidgetParams),
	}

	if len(w.params.Listen) == 0 {
		return nil, errors.New("missing 'listen'")
	}

	if len(w.params.Path) == 0 {
		return nil, errors.New("missing 'path'")
	}

	if instances == nil {
		instances = make(map[string]*httpInstance, 1)
	}

	if instance, ok := instances[w.params.Listen]; ok {
		if _, ok := instance.paths[w.params.Path]; ok {
			return nil, fmt.Errorf("path '%s' already in use", w.params.Path)
		}

		instance.mux.HandleFunc(w.params.Path, w.httpHandler)
		instance.paths[w.params.Path] = struct{}{}
	} else {
		instance := &httpInstance{
			mux:   http.NewServeMux(),
			paths: make(map[string]struct{}, 1),
		}
		instance.mux.HandleFunc(w.params.Path, w.httpHandler)
		instance.paths[w.params.Path] = struct{}{}
		instances[w.params.Listen] = instance

		w.httpServer = &http.Server{
			Addr:    w.params.Listen,
			Handler: instance.mux,
		}
	}

	return w, nil
}

// Run starts the main loop.
func (w *HTTPWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c
	if w.httpServer == nil {
		return nil
	}

	return w.httpServer.ListenAndServe()
}

// Event processes the widget events.
func (w *HTTPWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	if w.conn != nil {
		return websocket.JSON.Send(w.conn, event)
	}

	return nil
}

func (w *HTTPWidget) httpHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		ws := websocket.Handler(w.wsHandler)
		ws.ServeHTTP(response, request)

		return
	}

	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("%s", err)
		}

		var messages []ygs.I3BarBlock
		if err := json.Unmarshal(body, &messages); err != nil {
			log.Printf("%s", err)
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, "%s", err)
		}

		w.c <- messages

		return
	}

	response.WriteHeader(http.StatusBadRequest)

	_, err := response.Write([]byte("bad request method, allow GET for websocket and POST for HTTP update"))
	if err != nil {
		log.Printf("failed to write response: %s", err)
	}
}

func (w *HTTPWidget) wsHandler(ws *websocket.Conn) {
	var messages []ygs.I3BarBlock

	w.conn = ws

	for {
		if err := websocket.JSON.Receive(ws, &messages); err != nil {
			if err == io.EOF {
				if w.conn == ws {
					w.c <- nil
					w.conn = nil
				}

				break
			}

			log.Printf("%s", err)
		}

		if w.conn != ws {
			break
		}

		w.c <- messages
	}

	ws.Close()
}
