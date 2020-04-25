package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	ReceivedPackets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fluent_receive_msgs_total",
		Help: "Logging entries, incoming",
	})
	ReceivedBytes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fluent_receive_bytes_total",
		Help: "Logging traffic, incoming, bytes",
	})

	ForwardPackets = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "fluent_forward_msgs_total",
		Help: "Logging entries, forwarded",
	}, []string{"target"})
)

var Handler http.Handler

func init() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(ReceivedPackets, ReceivedBytes, ForwardPackets)

	promOpts := promhttp.HandlerOpts{EnableOpenMetrics: false, ErrorLog: log.StandardLogger()}
	Handler = promhttp.HandlerFor(registry, promOpts)
}