/*
Copyright 2025.

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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kirillyesikov/operator/api/v1"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/kirillyesikov/operator/api/v1"
)

// KirillAppReconciler reconciles a KirillApp object
type KirillAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=apps.kirillesikov.atwebpages.com,resources=kirillapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kirillesikov.atwebpages.com,resources=kirillapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kirillesikov.atwebpages.com,resources=kirillapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KirillApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *KirillAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	kirillApp := &appsv1.KirillApp{}
	err := r.Get(ctx, req.NamespacedName, kirillApp)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch KirillApp")
			return ctrl.Result{}, err
		}

		log.Info("KirillApp resource not found")
		return ctrl.Result{}, nil
	}

	err = r.ensureDeployment(ctx, kirillApp)
	if err != nil {
		log.Error(err, "unable to ensure Deployment for KirillApp")

		return ctrl.Result{}, err
	}

	err = r.ensureService(ctx, kirillApp)

	if err != nil {
		log.Error(err, "unable to ensure Service for KirillApp")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled KirillApp", "name", kirillApp.Name, "namespace", kirillApp.Namespace)
	return ctrl.Result{}, nil
}

func (r *KirillAppReconciler) ensureDeployment(ctx context.Context, kirillApp *v1.KirillApp) error {
	log := log.FromContext(ctx)

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name + "-deployment",
			Namespace: kirillApp.Namespace,
		},
		Spec: apps.DeploymentSpec{
			Replicas: pointer.Int32Ptr(2),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "radioapp-container",
							Image: "kyesikov/radio:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	err := r.Get(ctx, client.ObjectKey{Namespace: kirillApp.Namespace, Name: deployment.Name}, deployment)

	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Deployment for KirillApp")

			return err
		}

		log.Info("Creating a new Deployment for KirillApp", "name", deployment.Name)
		err := r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "failed to create Deployment for KirillApp")
			return err
		}
		log.Info("Deployment created successfully", "name", deployment.Name)
	} else {
		log.Info("Deployment already exists", "name", deployment.Name)

	}
	return nil
}

func (r *KirillAppReconciler) ensureService(ctx context.Context, kirillApp *appsv1.KirillApp) error {
	log := log.FromContext(ctx)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name + "-service",
			Namespace: kirillApp.Namespace,
		},
	}

	err := r.Get(ctx, client.ObjectKey{Namespace: kirillApp.Namespace, Name: service.Name}, service)

	if err != nil {

		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Service for KirillApp")
			return err
		}

		log.Info("Creating a new Service for KirillApp", "name", service.Name)

		err := r.Create(ctx, service)
		if err != nil {
			log.Error(err, "failed to create Service for KirillApp")
			return err
		}
		log.Info("Service created", "name", service.Name)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KirillAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.KirillApp{}).
		Complete(r)
}
