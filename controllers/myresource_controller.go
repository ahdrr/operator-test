/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	mygroupv1alpha1 "local.dev/myoperator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//+kubebuilder:rbac:groups=mygroup.local.dev,resources=myresources,verbs=get;list;watch;create;update;patch;delete
// 这个注解赋予了控制器对 "mygroup.local.dev" 组下的 "myresources" 资源执行 get、list、watch、create、update、patch 和 delete 操作的权限。

//+kubebuilder:rbac:groups=mygroup.local.dev,resources=myresources/status,verbs=get;update;patch
// 这个注解赋予了控制器对 "mygroup.local.dev" 组下的 "myresources" 资源的 "status" 子资源执行 get、update 和 patch 操作的权限。

//+kubebuilder:rbac:groups=mygroup.local.dev,resources=myresources/finalizers,verbs=update
// 这个注解赋予了控制器对 "mygroup.local.dev" 组下的 "myresources" 资源的 "finalizers" 子资源执行 update 操作的权限。

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// 这个注解赋予了控制器对 "apps" 组下的 "deployments" 资源执行 get、list、watch、create、update、patch 和 delete 操作的权限。

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// 这个注解赋予了控制器对 "core" 组下的 "services" 资源执行 get、list、watch、create、update、patch 和 delete 操作的权限。

//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// 事件events 创建和更新相关权限

// MyResourceReconciler reconciles a MyResource object
type MyResourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder

	//metricsCollector *metrics.MetricsCollector
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//1、设置context日志
	r.setLoger(ctx, req)

	//2、检查cr
	myResource, err := r.getCr(ctx, req)
	if err != nil {
		r.Log.Error(err, "Failed get cr")
		return ctrl.Result{}, err
	}
	//3、创建一个跟踪一组goroutine的错误组
	g, ctx := errgroup.WithContext(ctx)
	//4、后台创建或更新 Deployment
	g.Go(func() error { return r.ensureDeployment(ctx, myResource) })

	//5、后台创建或更新 services
	g.Go(func() error { return r.ensureServices(ctx, myResource) })

	//6、等待所有任务完成
	if err := g.Wait(); err != nil {
		r.Recorder.Event(myResource, corev1.EventTypeNormal, "error", err.Error())
		r.Log.Error(err, "Failed to ensure Resource.")

		return ctrl.Result{}, err
	}

	//5. 关联 Annotations
	//6、 更新关联资源
	// TODO(user): your logic here
	return ctrl.Result{}, nil

}

func (r *MyResourceReconciler) setLoger(ctx context.Context, req ctrl.Request) {
	r.Log = log.FromContext(ctx)

	//r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)

	r.Log.Info("===============Reconciling MyResource===================")
}

func (r *MyResourceReconciler) getCr(ctx context.Context, req ctrl.Request) (*mygroupv1alpha1.MyResource, error) {
	myResource := &mygroupv1alpha1.MyResource{}
	if err := r.Get(ctx, req.NamespacedName, myResource); err != nil {
		if k8sError.IsNotFound(err) {
			// 当 CR 被删除时，也会触发 Reconciliation。如果 CR 已被删除，我们只需要返回即可
			return nil, errors.New("MyResource not found. Ignoring since object must have been deleted")
		}
		// 其他错误
		return nil, err
	}
	//CR正在被删除
	if myResource.DeletionTimestamp != nil {
		return nil, errors.New("CR have been deleteding")
	}
	return myResource, nil
}

//func (r *MyResourceReconciler) boundAnnotations(ctx context.Context, mr *mygroupv1alpha1.MyResource) error {
//	// ...existing code to create or update Deployment...
//	// Check if the spec has changed
//	oldSpecData, oldSpecExists := mr.Annotations["spec"]
//	newSpecData, _ := json.Marshal(mr.Spec)
//	if !oldSpecExists || oldSpecData != string(newSpecData) {
//		// The spec has changed, update the annotation
//		if mr.Annotations == nil {
//			mr.Annotations = make(map[string]string)
//		}
//		mr.Annotations["spec"] = string(newSpecData)
//		if err := r.Update(ctx, mr); err != nil {
//			r.log.Error(err, "Failed to update MyResource annotations.")
//			return err
//		}
//	}
//	return nil
//}

type updateOrReplacePredicate struct {
	Log logr.Logger

	predicate.Funcs
}

func (e *updateOrReplacePredicate) Update(event.UpdateEvent) bool {
	// Update event
	e.Log.Info("UpdateEvent*************************")
	return false
}

func (e *updateOrReplacePredicate) Create(event.CreateEvent) bool {
	// Create event
	e.Log.Info("CreateEvent*************************")
	return false
}

func (e *updateOrReplacePredicate) Delete(event.DeleteEvent) bool {
	// Delete event
	e.Log.Info("DeleteEvent*************************")

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1alpha1.MyResource{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &mygroupv1alpha1.MyResource{},
		},
			builder.WithPredicates(&updateOrReplacePredicate{Log: r.Log}),
		).
		//Watches(
		//	&source.Kind{Type: &appsv1.Deployment{}},
		//	&handler.EnqueueRequestForOwner{
		//		IsController: true,
		//		OwnerType:    &mygroupv1alpha1.MyResource{},
		//	}).
		Complete(r)
}
