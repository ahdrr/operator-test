package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//如果开启自定义metrics,需要添加一下面代码，集成operator Metrics模式
// 在Reconcile中初始化MetricsCollector
//if r.metricsCollector == nil {
//	r.metricsCollector = metrics.NewMetricsCollector(r.getPodList(metav1.LabelSelector{}))
//	go r.collectMetrics()
//}

func (r *MyResourceReconciler) collectMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		//r.metricsCollector.Collect()
	}
}

// getPodList函数，获取带有特定标签的Pod的IP地址
func (r *MyResourceReconciler) getPodList(labelSelector metav1.LabelSelector) (func() ([]string, error), error) {

	// 首先，将metav1.LabelSelector转换为labels.Selector
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to convert labelSelector: %w", err)
	}

	return func() ([]string, error) {

		// 用于存储Pod IP的切片
		var podIPs []string

		// 创建PodList对象
		podList := &corev1.PodList{}

		// 根据标签选择器获取Pod
		if err := r.List(context.Background(), podList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
			return nil, err
		}

		// 遍历Pod列表，获取每个Pod的IP
		for _, pod := range podList.Items {
			podIPs = append(podIPs, pod.Status.PodIP)
		}

		return podIPs, nil
	}, nil
}
