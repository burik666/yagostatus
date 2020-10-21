package config

import (
	"encoding/json"

	"github.com/burik666/yagostatus/ygs"
)

// ErrorWidget creates new widget with error message.
func ErrorWidget(text string) WidgetConfig {
	blocks, _ := json.Marshal([]ygs.I3BarBlock{
		{
			FullText: text,
			Color:    "#ff0000",
		},
	})

	return WidgetConfig{
		Name: "static",
		Params: map[string]interface{}{
			"blocks": string(blocks),
		},
		File: "builtin",
	}
}
