package metrics

// 你的控制器代码所在的包，需要使用你的实际路径替换

type MetricsCollector struct {
	podListFunc func() ([]string, error) // 用于获取pod列表的函数
}

func NewMetricsCollector(podListFunc func() ([]string, error)) *MetricsCollector {
	return &MetricsCollector{
		podListFunc: podListFunc,
	}
}

func (mc *MetricsCollector) Collect() {
	podList, err := mc.podListFunc()
	if err != nil {
		// log error
		return
	}

	for _, podIP := range podList {
		stats, err := getNginxStats(podIP)
		if err != nil {
			// log error
			continue
		}

		// 按你的需求从stats中提取数据并发送到channel
		TcpConnections.WithLabelValues(podIP).Set(float64(stats["active"]))
		// 更多metrics...
	}
}

func getNginxStats(podIP string) (map[string]float64, error) {
	return map[string]float64{
		"active":   100,
		"reading":  10,
		"writing":  20,
		"waiting":  30,
		"accepted": 200,
		"handled":  190,
		"requests": 1000,
	}, nil
}
