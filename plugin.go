package prometheus

import (
	"github.com/VolantMQ/vlapi/vlplugin"
)

type prometheusPlugin struct {
	vlplugin.Descriptor
}

var _ vlplugin.Plugin = (*prometheusPlugin)(nil)
var _ vlplugin.Info = (*prometheusPlugin)(nil)

// Plugin symbol
var Plugin prometheusPlugin

var version string

func init() {
	Plugin.V = version
	Plugin.N = "prometheus"
	Plugin.T = "monitoring"
}

// Info plugin info
func (pl *prometheusPlugin) Info() vlplugin.Info {
	return pl
}
