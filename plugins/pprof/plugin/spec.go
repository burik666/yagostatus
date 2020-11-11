package plugin

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/burik666/yagostatus/ygs"
)

type Params struct {
	Listen string
}

var srv *http.Server

var Spec = ygs.PluginSpec{
	Name: "pprof",
	DefaultParams: Params{
		Listen: "localhost:6060",
	},
	InitFunc: func(p interface{}, l ygs.Logger) error {
		params := p.(Params)
		l.Infof("http://%s/debug/pprof", params.Listen)

		srv = &http.Server{
			Addr: params.Listen,
		}

		mux := http.ServeMux{}

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		go func() {
			l.Infof("%s", srv.ListenAndServe())
		}()

		return nil
	},
	ShutdownFunc: func() error {
		if srv == nil {
			return nil
		}

		return srv.Shutdown(context.Background())
	},
}
