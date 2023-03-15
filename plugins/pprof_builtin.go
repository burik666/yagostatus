// +build plugin_pprof

package plugins

import (
	"github.com/burik666/yagostatus/plugins/pprof/plugin"
	"github.com/burik666/yagostatus/ygs"
)

func init() {
	if err := ygs.RegisterPlugin(plugin.Spec); err != nil {
		panic(err)
	}
}
