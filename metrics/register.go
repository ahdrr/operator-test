package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// 定义你的指标
	TcpConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nginx_tcp_connections",
			Help: "Number of TCP connections",
		},
		[]string{"pod"},
	)
)

func init() {
	// 注册你的指标
	prometheus.MustRegister(TcpConnections)
}
