/*
Copyright (c) 2022 maxcleme

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controllers

import (
	"context"

	dockercomv1alpha1 "github.com/maxcleme/compose-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ComposeReconciler reconciles a Compose object
type ComposeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=docker.com,resources=composes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=docker.com,resources=composes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=docker.com,resources=composes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Compose object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ComposeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Memcached instance
	compose := &dockercomv1alpha1.Compose{}
	err := r.Get(ctx, req.NamespacedName, compose)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Compose resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Compose")
		return ctrl.Result{}, err
	}

	log.Info("compose", "spec", compose.Spec.Spec)

	//// Check if the deployment already exists, if not create a new one
	//found := &appsv1.Deployment{}
	//err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)
	//if err != nil && errors.IsNotFound(err) {
	//	// Define a new deployment
	//	dep := r.deploymentForMemcached(memcached)
	//	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	//	err = r.Create(ctx, dep)
	//	if err != nil {
	//		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	//		return ctrl.Result{}, err
	//	}
	//	// Deployment created successfully - return and requeue
	//	return ctrl.Result{Requeue: true}, nil
	//} else if err != nil {
	//	log.Error(err, "Failed to get Deployment")
	//	return ctrl.Result{}, err
	//}
	//
	//// Ensure the deployment size is the same as the spec
	//size := memcached.Spec.Size
	//if *found.Spec.Replicas != size {
	//	found.Spec.Replicas = &size
	//	err = r.Update(ctx, found)
	//	if err != nil {
	//		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	//		return ctrl.Result{}, err
	//	}
	//	// Ask to requeue after 1 minute in order to give enough time for the
	//	// pods be created on the cluster side and the operand be able
	//	// to do the next update step accurately.
	//	return ctrl.Result{RequeueAfter: time.Minute}, nil
	//}
	//
	//// Update the Memcached status with the pod names
	//// List the pods for this memcached's deployment
	//podList := &corev1.PodList{}
	//listOpts := []client.ListOption{
	//	client.InNamespace(memcached.Namespace),
	//	client.MatchingLabels(labelsForMemcached(memcached.Name)),
	//}
	//if err = r.List(ctx, podList, listOpts...); err != nil {
	//	log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
	//	return ctrl.Result{}, err
	//}
	//podNames := getPodNames(podList.Items)
	//
	//// Update status.Nodes if needed
	//if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
	//	memcached.Status.Nodes = podNames
	//	err := r.Status().Update(ctx, memcached)
	//	if err != nil {
	//		log.Error(err, "Failed to update Memcached status")
	//		return ctrl.Result{}, err
	//	}
	//}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComposeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dockercomv1alpha1.Compose{}).
		Complete(r)
}
