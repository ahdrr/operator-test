package controllers

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"

	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	mygroupv1alpha1 "local.dev/myoperator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MyResourceReconciler) newServices(mr *mygroupv1alpha1.MyResource) []*corev1.Service {
	appService := r.createService(mr, "-service", mr.Spec.Service.Type, mr.Spec.Service.Ports)

	services := []*corev1.Service{appService}

	if mr.Spec.Monitoring != nil && mr.Spec.Monitoring.Enabled {
		monitoringPorts := []corev1.ServicePort{
			{
				Port:       mr.Spec.Monitoring.ExporterPort,
				TargetPort: intstr.FromInt(int(mr.Spec.Monitoring.ExporterPort)),
			},
		}
		monitoringService := r.createService(mr, "-monitoring-service",
			corev1.ServiceTypeClusterIP, monitoringPorts)
		services = append(services, monitoringService)
	}

	return services
}

func (r *MyResourceReconciler) createService(mr *mygroupv1alpha1.MyResource, serviceNameSuffix string, serviceType corev1.ServiceType, ports []corev1.ServicePort) *corev1.Service {
	serviceName := mr.Name + serviceNameSuffix
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: mr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": mr.Name + "-deployment",
			},
			Type:  corev1.ServiceType(serviceType),
			Ports: ports,
		},
	}
}

func (r *MyResourceReconciler) ensureServices(ctx context.Context, mr *mygroupv1alpha1.MyResource) error {
	services := r.newServices(mr)

	for _, service := range services {
		if err := r.ensureSingleService(ctx, service, mr); err != nil {
			return err
		}
	}
	return nil
}

func (r *MyResourceReconciler) ensureSingleService(ctx context.Context, service *corev1.Service, mr *mygroupv1alpha1.MyResource) error {
	// Set MyResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(mr, service, r.Scheme); err != nil {
		return err
	}

	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: mr.Namespace}, found)
	if err != nil {
		if k8sError.IsNotFound(err) {
			// Service not found, create it
			r.Log.Info("Creating new Service", "Service.Name", service.Name)
			err = r.Create(ctx, service)
			if err != nil {
				return fmt.Errorf("failed to create new service +%s: %w", service.Name, err)
			}
			r.Recorder.Event(mr, corev1.EventTypeNormal, "Created", fmt.Sprintf("Created service %s", service.Name))
		} else {
			// Other errors
			return err
		}
	} else {
		if reflect.DeepEqual(found.Spec, service.Spec) {
			return nil
		}
		r.Log.Info("Updating existing Service " + found.Name)
		found.Spec = service.Spec
		if err = r.Update(ctx, found); err != nil {
			return fmt.Errorf("failed to update existing service +%s: %w", found.Name, err)

		}
		r.Recorder.Event(mr, corev1.EventTypeNormal, "Update", fmt.Sprintf("Update service %s", service.Name))
	}
	return nil
}
