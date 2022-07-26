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
	"fmt"

	spec "github.com/compose-spec/compose-go/loader"
	composeTypes "github.com/compose-spec/compose-go/types"
	dockercomv1alpha1 "github.com/maxcleme/compose-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

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

	c := &dockercomv1alpha1.Compose{}
	err := r.Get(ctx, req.NamespacedName, c)
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

	project, err := spec.Load(composeTypes.ConfigDetails{
		ConfigFiles: []composeTypes.ConfigFile{
			{Content: []byte(c.Spec.Spec)},
		},
		Environment: map[string]string{},
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// build service index for later check as we are already going to iterate over all services
	serviceIndex := make(map[string]struct{}, len(project.Services))
	for _, s := range project.Services {
		log.Info("compose", "service", s)
		serviceIndex[s.Name] = struct{}{}

		// Define a new deployment
		dep, err := r.deploymentForService(c, project, s)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Check if the deployment already exists, if not create a new one
		found := &appsv1.Deployment{}
		err = r.Get(ctx, types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				log.Error(err, "Failed to create Deployment")
				return ctrl.Result{}, err
			}
			continue
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		// otherwise, patch it
		log.Info("Patching Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		if err := r.Update(ctx, dep); err != nil {
			log.Error(err, "Failed to patch Deployment")
			return ctrl.Result{}, err
		}
	}

	list := &appsv1.DeploymentList{}
	listOpts := []client.ListOption{
		client.InNamespace(c.Namespace),
		client.MatchingLabels{"project": project.Name},
	}
	if err = r.List(ctx, list, listOpts...); err != nil {
		log.Error(err, "Failed to list all Deployments")
		return ctrl.Result{}, err
	}

	// check if some are missing, implying service deletion
	for _, dep := range list.Items {
		if _, ok := serviceIndex[dep.Labels["service"]]; !ok {
			// service has been deleted
			log.Info("Deleting Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			if err := r.Delete(ctx, &dep); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComposeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dockercomv1alpha1.Compose{}).
		Complete(r)
}

func labelsForService(p *composeTypes.Project, s composeTypes.ServiceConfig) map[string]string {
	return map[string]string{"project": p.Name, "service": s.Name}
}

func (r *ComposeReconciler) deploymentForService(c *dockercomv1alpha1.Compose, p *composeTypes.Project, s composeTypes.ServiceConfig) (*appsv1.Deployment, error) {
	ls := labelsForService(p, s)

	replicas := int32(1)
	if s.Deploy != nil && s.Deploy.Replicas != nil {
		replicas = int32(*s.Deploy.Replicas)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("compose-%s-%s", p.Name, s.Name),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: s.Image,
						Name:  s.Name,
					}},
				},
			},
		},
	}

	return dep, ctrl.SetControllerReference(c, dep, r.Scheme)
}
