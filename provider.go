package prometheus

import (
	"fmt"
	"strings"
	"time"

	"github.com/VolantMQ/vlapi/vlmonitoring"
	"github.com/VolantMQ/vlapi/vlplugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultPath = "/metrics"
)

// Config to specificy prometheus metrics endpoint
type Config struct {
	Path string `mapstructure:"path,omitempty" yaml:"path,omitempty" json:"path,omitempty" default:""`
	Port string `mapstructure:"port,omitempty" yaml:"port,omitempty" json:"port,omitempty" default:""`
}

type flowCounter struct {
	sent prometheus.Counter
	recv prometheus.Counter
}

type flowGauge struct {
	sent prometheus.Gauge
	recv prometheus.Gauge
}

type clients struct {
	connected prometheus.Gauge
	persisted prometheus.Gauge
	total     prometheus.Gauge
	maximum   prometheus.Gauge
}

type subscriptions struct {
	total   prometheus.Gauge
	maximum prometheus.Gauge
}

type packets struct {
	*flowCounter
	connect    prometheus.Counter
	connack    prometheus.Counter
	publish    *flowCounter
	puback     *flowCounter
	pubrec     *flowCounter
	pubrel     *flowCounter
	pubcomp    *flowCounter
	sub        prometheus.Counter
	suback     prometheus.Counter
	unsub      prometheus.Counter
	unsuback   prometheus.Counter
	pingreq    prometheus.Counter
	pingresp   prometheus.Counter
	disconnect *flowCounter
	auth       *flowCounter
	unknown    prometheus.Counter
	inflight   *flowGauge
	retained   prometheus.Gauge
}

type server struct {
	startTS time.Time
	uptime  prometheus.Gauge
}

type stat struct {
	server  *server
	bytes   *flowCounter
	clients *clients
	packets *packets
	subs    *subscriptions
}

type impl struct {
	*vlplugin.SysParams
	config Config
	stat   *stat
}

var _ vlmonitoring.IFace = (*impl)(nil)

func (pl *prometheusPlugin) Load(c interface{}, params *vlplugin.SysParams) (interface{}, error) {

	cfg := c.(Config)

	im := &impl{
		SysParams: params,
		config:    cfg,
	}

	if im.config.Path == "" {
		im.config.Path = defaultPath
	}

	im.config.Path = strings.TrimSuffix(im.config.Path, "/")

	im.stat = &stat{
		server: newServer(),
		bytes: newFlowCounter(
			"mqtt_bytes_total_sent",
			"The total number of bytes sent since server start",
			"mqtt_bytes_total_recv",
			"The total number of bytes received since server start"),
		clients: newClients(),
		packets: newPackets(),
		subs:    newSubscriptions(),
	}

	params.GetHTTPServer(im.config.Port).Mux().Handle(im.config.Path, promhttp.Handler())

	return im, nil
}

func (im *impl) Push(stats vlmonitoring.Stats) {
	im.stat.server.uptime.Set(time.Since(im.stat.server.startTS).Seconds())

	im.stat.bytes.sent.Add(float64(stats.Bytes.Sent.Get()))
	im.stat.bytes.recv.Add(float64(stats.Bytes.Recv.Get()))

	im.stat.clients.connected.Set(float64(stats.Clients.Connected.Get()))
	im.stat.clients.persisted.Set(float64(stats.Clients.Persisted.Get()))
	im.stat.clients.total.Set(float64(stats.Clients.Total.Get()))
	im.stat.clients.maximum.Set(float64(stats.Clients.Total.Max))

	im.stat.packets.sent.Add(float64(stats.Packets.Total.Sent.Get()))
	im.stat.packets.recv.Add(float64(stats.Packets.Total.Recv.Get()))
	im.stat.packets.connect.Add(float64(stats.Packets.Connect.Get()))
	im.stat.packets.connack.Add(float64(stats.Packets.ConnAck.Get()))
	im.stat.packets.publish.sent.Add(float64(stats.Packets.Publish.Sent.Get()))
	im.stat.packets.publish.recv.Add(float64(stats.Packets.Publish.Recv.Get()))
	im.stat.packets.puback.sent.Add(float64(stats.Packets.Puback.Sent.Get()))
	im.stat.packets.puback.recv.Add(float64(stats.Packets.Puback.Recv.Get()))
	im.stat.packets.pubrec.sent.Add(float64(stats.Packets.Pubrec.Sent.Get()))
	im.stat.packets.pubrec.recv.Add(float64(stats.Packets.Pubrec.Recv.Get()))
	im.stat.packets.pubrel.sent.Add(float64(stats.Packets.Pubrel.Sent.Get()))
	im.stat.packets.pubrel.recv.Add(float64(stats.Packets.Pubrel.Recv.Get()))
	im.stat.packets.pubcomp.sent.Add(float64(stats.Packets.Pubcomp.Sent.Get()))
	im.stat.packets.pubcomp.recv.Add(float64(stats.Packets.Pubcomp.Recv.Get()))
	im.stat.packets.sub.Add(float64(stats.Packets.Sub.Get()))
	im.stat.packets.suback.Add(float64(stats.Packets.SubAck.Get()))
	im.stat.packets.unsub.Add(float64(stats.Packets.UnSub.Get()))
	im.stat.packets.unsuback.Add(float64(stats.Packets.UnSubAck.Get()))
	im.stat.packets.pingreq.Add(float64(stats.Packets.PingReq.Get()))
	im.stat.packets.pingresp.Add(float64(stats.Packets.PingResp.Get()))
	im.stat.packets.disconnect.sent.Add(float64(stats.Packets.Disconnect.Sent.Get()))
	im.stat.packets.disconnect.recv.Add(float64(stats.Packets.Disconnect.Recv.Get()))
	im.stat.packets.auth.sent.Add(float64(stats.Packets.Auth.Sent.Get()))
	im.stat.packets.auth.recv.Add(float64(stats.Packets.Auth.Recv.Get()))

	im.stat.packets.unknown.Add(float64(stats.Packets.Unknown.Get()))

	im.stat.packets.inflight.sent.Set(float64(stats.Packets.UnAckSent.Get()))
	im.stat.packets.inflight.recv.Set(float64(stats.Packets.UnAckRecv.Get()))

	im.stat.packets.retained.Set(float64(stats.Packets.Retained.Get()))

	im.stat.subs.total.Set(float64(stats.Subs.Total.Get()))
	im.stat.subs.maximum.Set(float64(stats.Subs.Total.Max))
}

func (im *impl) Shutdown() error {
	return nil
}

func newServer() *server {
	s := &server{
		startTS: time.Now(),
		uptime: newGauge(
			"mqtt_server_uptime",
			"The amount of seconds since server started"),
	}

	return s
}

func newClients() *clients {
	c := &clients{
		connected: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_clients_connected",
			Help: "The number of clients connected to the server",
		}),
		persisted: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_clients_persisted",
			Help: "The number of clients persisted on server",
		}),
		total: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_clients_total",
			Help: "The total number of active and inactive clients",
		}),
		maximum: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "mqtt_clients_maximum",
			Help: "The number of clients ever connected to the server",
		}),
	}

	prometheus.MustRegister(c.connected)
	prometheus.MustRegister(c.persisted)
	prometheus.MustRegister(c.total)
	prometheus.MustRegister(c.maximum)

	return c
}

func newCounter(n, h string) prometheus.Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: n,
		Help: h,
	})

	prometheus.MustRegister(c)

	return c
}

func newGauge(n, h string) prometheus.Gauge {
	c := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: n,
		Help: h,
	})

	prometheus.MustRegister(c)

	return c
}

func newFlowCounter(n1, h1, n2, h2 string) *flowCounter {
	f := &flowCounter{
		sent: newCounter(n1, h1),
		recv: newCounter(n2, h2),
	}

	return f
}

func newFlowGauge(n1, h1, n2, h2 string) *flowGauge {
	f := &flowGauge{
		sent: newGauge(n1, h1),
		recv: newGauge(n2, h2),
	}

	return f
}

func newSubscriptions() *subscriptions {
	s := &subscriptions{
		total: newGauge(
			"mqtt_subscriptions_total",
			"The total number of active subscriptions"),
		maximum: newGauge(
			"mqtt_subscriptions_maximum",
			"The maximum amount active subscriptions has ever been on server since start"),
	}

	return s
}

func newPackets() *packets {
	formatHints := func(n string) (string, string, string, string) {
		return fmt.Sprintf("mqtt_packets_%s_sent", n),
			fmt.Sprintf("The total number of %s packets sent since server start", n),
			fmt.Sprintf("mqtt_packets_%s_recv", n),
			fmt.Sprintf("The total number of %s packets recv since server start", n)
	}
	p := &packets{
		flowCounter: newFlowCounter(
			"mqtt_packets_total_sent",
			"The total number of packets of all types sent since server start",
			"mqtt_packets_total_recv",
			"The total number of packets of all types received since server start",
		),
		connect:    newCounter("mqtt_packets_connect_recv", "The total number of connect packets recv since server start"),
		connack:    newCounter("mqtt_packets_connack_recv", "The total number of connack packets sent since server start"),
		publish:    newFlowCounter(formatHints("publish")),
		puback:     newFlowCounter(formatHints("puback")),
		pubrec:     newFlowCounter(formatHints("pubrec")),
		pubrel:     newFlowCounter(formatHints("pubrel")),
		pubcomp:    newFlowCounter(formatHints("pubcomp")),
		sub:        newCounter("mqtt_packets_subscribe_recv", "The total number of subscribe packets recv since server start"),
		suback:     newCounter("mqtt_packets_suback_recv", "The total number of suback packets sent since server start"),
		unsub:      newCounter("mqtt_packets_unsubscribe_recv", "The total number of unsubscribe packets recv since server start"),
		unsuback:   newCounter("mqtt_packets_unsuback_recv", "The total number of unsuback packets sent since server start"),
		pingreq:    newCounter("mqtt_packets_pingreq_recv", "The total number of pingreq packets recv since server start"),
		pingresp:   newCounter("mqtt_packets_pingresp_recv", "The total number of pingresp packets sent since server start"),
		disconnect: newFlowCounter(formatHints("disconnect")),
		auth:       newFlowCounter(formatHints("auth")),
		unknown:    newCounter("mqtt_packets_unknown_recv", "The total number of unknown packets recv since server start"),
		inflight: newFlowGauge(
			"mqtt_packets_inflight_sent",
			"The total number of packets awaiting for client acknowledgement",
			"mqtt_packets_inflight_recv",
			"The total number of packets awaiting for server acknowledgement",
		),
		retained: newGauge("mqtt_packets_retained", "The total number of retained packets"),
	}

	return p
}
