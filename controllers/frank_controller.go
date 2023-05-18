/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	appsv1 "github.com/jianlong0808/operator-frank/api/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// FrankReconciler reconciles a Frank object
type FrankReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//K8s 为我们提供了2种等级的 event，分别是 Normal 和 Warning。
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.frank.com,resources=franks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.frank.com,resources=franks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.frank.com,resources=franks/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Frank object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *FrankReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	Log := log.FromContext(ctx)
	frank := &appsv1.Frank{}
	deployment := &k8sappsv1.Deployment{}
	//获取frank
	err := r.Get(ctx, req.NamespacedName, frank)
	//忽略掉 not-found 错误，它们不能通过重新排队修复（要等待新的通知）
	//在删除一个不存在的对象时，可能会报这个错误。
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	frankFinalizerName := "finalizer.frank.com"

	//判断frank对象是否被删除,如果值为0,则表示frank对象没有在删除, 此时就需要将finalizer添加到frank对象中
	//如果值为非0,则表示frank对象正在被删除,此时应该执行资源清理操作, 如果清理成功, 就将finalizer从frank对象中删除, 删除finalier字段后资源会自动清理
	if frank.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(frank.ObjectMeta.Finalizers, frankFinalizerName) {
			Log.Info("添加 Finalizer")
			frank.ObjectMeta.Finalizers = append(frank.ObjectMeta.Finalizers, frankFinalizerName)
			if err := r.Update(ctx, frank); err != nil {
				Log.Error(err, "添加 Finalizer 失败")
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(frank.ObjectMeta.Finalizers, frankFinalizerName) {
			Log.Info("回调 Delete 资源成功")
			if err := r.deleteHook(frank); err != nil {
				//如果删除失败, 10秒后重试
				Log.Error(err, "回调 Delete 资源失败")
				return ctrl.Result{RequeueAfter: time.Second * 10}, err
			}
			frank.ObjectMeta.Finalizers = removeString(frank.ObjectMeta.Finalizers, frankFinalizerName)
			if err := r.Update(ctx, frank); err != nil {
				Log.Error(err, "删除 Finalizer 失败")
				//如果删除失败, 10秒后重试
				return ctrl.Result{RequeueAfter: time.Second * 10}, err
			}
		}
	}

	//获取deployment
	err = r.Get(ctx, req.NamespacedName, deployment)

	if err != nil {
		if errors.IsNotFound(err) {
			Log.Info("Deployment Not Found")
			err = r.CreateDeployment(ctx, frank)
			if err != nil {
				Log.Error(err, "Create Deployment Failed")
				r.Recorder.Event(frank, k8scorev1.EventTypeWarning, "FailedCreateDeployment", err.Error())
				return ctrl.Result{}, err
			}
			Log.Info("Create Deployment Success")
			Log.Info("Update Status")
			frank.Status.RealReplica = *frank.Spec.Replica
			//必须使用UpdateStatus方法来更新状态, 否则不会生效(describe或者 -o yaml 的时候不会显示)
			if err := r.Status().Update(ctx, frank); err != nil {
				Log.Error(err, "FailedUpdateStatus")
				r.Recorder.Event(frank, k8scorev1.EventTypeWarning, "FailedUpdateStatus", err.Error())
				return ctrl.Result{}, err
			}
			Log.Info("Update Status Success")
		} else {
			Log.Error(err, "Else Error, Not IsNotFound")
		}
		return ctrl.Result{}, err
	}

	//当frank对象的spec副本数发生变化的时候调整子资源deploy的副本数
	if *frank.Spec.Replica != *deployment.Spec.Replicas {
		Log.Info("Set Deploy's Replicas")
		deployment.Spec.Replicas = frank.Spec.Replica
		if err := r.Update(ctx, deployment); err != nil {
			Log.Error(err, "Update Deployment Failed")
			return ctrl.Result{}, err
		}
	}
	//当frank对象的spec镜像发生变化的时候触发deploy资资源的镜像更新
	if *frank.Spec.Image != deployment.Spec.Template.Spec.Containers[0].Image {
		Log.Info("Set Deploy's Image")
		deployment.Spec.Template.Spec.Containers[0].Image = *frank.Spec.Image
		if err := r.Update(ctx, deployment); err != nil {
			Log.Error(err, "Update Deployment Failed")
			return ctrl.Result{}, err
		}
	}

	frank.Status.RealReplica = *frank.Spec.Replica
	Log.Info("Update Status")
	//必须使用UpdateStatus方法来更新状态, 否则不会生效(describe或者 -o yaml 的时候不会显示)
	if err := r.Status().Update(ctx, frank); err != nil {
		//记录Event事件, describe的时候会看到, 如果err不为空, 触发下一次调谐
		r.Recorder.Event(frank, k8scorev1.EventTypeWarning, "FailedUpdateStatus", err.Error())
		return ctrl.Result{}, err
	}
	Log.Info("Update Status Success")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// 将 Reconcile 添加到 manager 中，这样当 manager 启动时它就会被启动, 如果我们想只关注由 CR 创建的 Deployment 因此我们可以采用 Owns() 方法
// 注: 也可以在这里对FrankReconciler进行一些初始化操作
func (r *FrankReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		For(&appsv1.Frank{}).
		Owns(&k8sappsv1.Deployment{}).
		Complete(r)
}

func (r *FrankReconciler) CreateDeployment(ctx context.Context, frank *appsv1.Frank) error {
	Log := log.FromContext(ctx)
	deployment := &k8sappsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: frank.Namespace,
			Name:      frank.Name,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(*frank.Spec.Replica),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": frank.Name,
				},
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": frank.Name,
					},
				},
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{
						{
							Name:            frank.Name,
							Image:           *frank.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							Ports: []k8scorev1.ContainerPort{
								{
									Name:          frank.Name,
									Protocol:      k8scorev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	/*
			必须先绑定资源父子关系, 再创建子资源, 否则删除crd的时候不会回收资资源, 为什么?
			因为: 通过SetControllerReference设置好父子关系后会在deployment加上类似于这样的信息以绑定父资源:
			ownerReferences:
			- apiVersion: apps.frank.com/v1
		      blockOwnerDeletion: true
			  controller: true
			  kind: Frank
			  name: frank-sample
			  uid: bc5f8da9-9f30-447b-8ae9-f6e409e378f6
	*/

	//binding deployment to frank
	if err := ctrl.SetControllerReference(frank, deployment, r.Scheme); err != nil {
		Log.Error(err, "SetControllerReference Failed")
		return err
	}

	if err := r.Create(ctx, deployment); err != nil {
		Log.Error(err, "Create Deployment Failed")
		return err
	}

	return nil
}

// containsString
// @Description: 判断slice中是否包含s
// @param slice
// @param s
// @return bool
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString
// @Description: 删除slice中的s
// @param slice
// @param s
// @return result
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return
}

// deleteHook
// @Description: 删除钩子
// @receiver r
// @param frank
// @return error
func (r *FrankReconciler) deleteHook(frank *appsv1.Frank) error {
	//TODO(user): your logic here
	return nil
}
