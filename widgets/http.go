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

// HTTPWidget implements the http server widget.
type HTTPWidget struct {
	c      chan<- []ygs.I3BarBlock
	conn   *websocket.Conn
	listen string
	path   string
}

var serveMuxes map[string]*http.ServeMux

// Configure configures the widget.
func (w *HTTPWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["listen"]
	if !ok {
		return errors.New("missing 'listen' setting")
	}
	w.listen = v.(string)

	v, ok = cfg["path"]
	if !ok {
		return errors.New("missing 'path' setting")
	}
	w.path = v.(string)

	if serveMuxes == nil {
		serveMuxes = make(map[string]*http.ServeMux, 1)
	}

	return nil
}

// Run starts the main loop.
func (w *HTTPWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c

	mux, ok := serveMuxes[w.listen]
	if ok {
		mux.HandleFunc(w.path, w.httpHandler)
		return nil
	}
	mux = http.NewServeMux()
	mux.HandleFunc(w.path, w.httpHandler)

	httpServer := &http.Server{
		Addr:    w.listen,
		Handler: mux,
	}
	serveMuxes[w.listen] = mux
	return httpServer.ListenAndServe()
}

// Event processes the widget events.
func (w *HTTPWidget) Event(event ygs.I3BarClickEvent) {
	if w.conn != nil {
		websocket.JSON.Send(w.conn, event)
	}

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
	response.Write([]byte("Bad request method, allow GET for websocket and POST for HTTP update"))
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

// Stop shutdowns the widget.
func (w *HTTPWidget) Stop() {}

func init() {
	ygs.RegisterWidget(&HTTPWidget{})
}
