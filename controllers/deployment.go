package controllers

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	mygroupv1alpha1 "local.dev/myoperator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MyResourceReconciler) newDeployment(mr *mygroupv1alpha1.MyResource) *appsv1.Deployment {
	// 创建要使用的Deployment对象的名称
	deploymentName := mr.Name + "-deployment"

	// 主容器定义
	containers := []corev1.Container{{
		Name:      deploymentName,
		Image:     mr.Spec.App.Image,
		Env:       mr.Spec.App.Envs,
		Resources: mr.Spec.App.Resources,
		Ports:     mr.Spec.App.Ports,
	}}

	// 如果提供了监控配置，添加一个监控Sidecar
	if mr.Spec.Monitoring != nil && mr.Spec.Monitoring.Enabled {
		sidecar := corev1.Container{
			Name:  "monitoring-sidecar",
			Image: mr.Spec.Monitoring.Image,
			Ports: []corev1.ContainerPort{{
				ContainerPort: mr.Spec.Monitoring.ExporterPort,
			}},
		}
		containers = append(containers, sidecar)
	}

	// 定义Deployment对象
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: mr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mr.Spec.App.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentName},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: mr.Spec.App.ImagePullSecrets,
					Containers:       containers,
				},
			},
		},
	}

}

func (r *MyResourceReconciler) ensureDeployment(ctx context.Context, mr *mygroupv1alpha1.MyResource) error {
	// 定义Deployment对象
	deploy := r.newDeployment(mr)
	// 设置MyResource实例为Deployment的所有者和控制器
	controllerutil.SetControllerReference(mr, deploy, r.Scheme)
	// 检查Deployment是否已经存在
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil {
		if k8sError.IsNotFound(err) {
			// Deployment不存在，创建新的Deployment
			r.Log.Info("Creating a new Deployment " + deploy.Name)
			err = r.Create(ctx, deploy)
			if err != nil {
				return fmt.Errorf("failed to create new Deployment: %w", err)
			}
			r.Recorder.Event(mr, corev1.EventTypeNormal, "Created", fmt.Sprintf("Created deployment %s", deploy.Name))
		} else {
			// 出现其他错误时，重新queue此项
			return fmt.Errorf("failed to get Deployment: %w", err)
		}
	} else {
		if reflect.DeepEqual(found.Spec, deploy.Spec) {
			r.Log.Info("Spec is unchanged")
			return nil
		}
		r.Log.Info("Updating existing Deployment", "Deployment.Namespace", found.Name)
		found.Spec = deploy.Spec
		if err = r.Update(ctx, found); err != nil {
			return fmt.Errorf("failed to update existing Deployment +%s: %w", found.Name, err)

		}
		r.Recorder.Event(mr, corev1.EventTypeNormal, "Updated", fmt.Sprintf("Updated deployment %s", deploy.Name))
	}
	return nil
}

//func newContainers(app *v1.AppService) []corev1.Container {
//	containerPorts := []corev1.ContainerPort{}
//	for _, svcPort := range app.Spec.Ports {
//		cport := corev1.ContainerPort{}
//		cport.ContainerPort = svcPort.TargetPort.IntVal
//		containerPorts = append(containerPorts, cport)
//	}
//	return []corev1.Container{
//		{
//			Name: app.Name,
//			Image: app.Spec.Image,
//			Resources: app.Spec.Resources,
//			Ports: containerPorts,
//			ImagePullPolicy: corev1.PullIfNotPresent,
//			Env: app.Spec.Envs,
//		},
//	}
//}
