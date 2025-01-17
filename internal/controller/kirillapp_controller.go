package controllers

import (
	context "context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myappv1 "apps.kirillesikov.atwebpages.com/kirillapp/api/v1" // Update the import path to match your project structure
)

// KirillAppReconciler reconciles a KirillApp object
type KirillAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=example.com,resources=kirillapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.com,resources=kirillapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.com,resources=kirillapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *KirillAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the KirillApp instance
	kirillApp := &myappv1.KirillApp{}
	if err := r.Get(ctx, req.NamespacedName, kirillApp); err != nil {
		logger.Error(err, "Failed to get KirillApp")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define a new Deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name,
			Namespace: kirillApp.Namespace,
		},
	}

	// Check if the Deployment already exists, if not, create it
	if err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

			// Set KirillApp instance as the owner and controller
			if err := controllerutil.SetControllerReference(kirillApp, deployment, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference")
				return ctrl.Result{}, err
			}

			// Create Deployment
			deployment = r.constructDeployment(kirillApp)
			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "Failed to create Deployment")
				return ctrl.Result{}, err
			}
		}
	} else {
		logger.Info("Deployment already exists", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
	}

	// Ensure the deployment matches the spec of KirillApp
	updatedDeployment := r.constructDeployment(kirillApp)
	if !deploymentEqual(deployment, updatedDeployment) {
		logger.Info("Updating Deployment to match KirillApp spec")
		deployment.Spec = updatedDeployment.Spec
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		}
	}

	// Define a new Service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name,
			Namespace: kirillApp.Namespace,
		},
	}

	// Check if the Service already exists, if not, create it
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, service); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)

			// Set KirillApp instance as the owner and controller
			if err := controllerutil.SetControllerReference(kirillApp, service, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference")
				return ctrl.Result{}, err
			}

			// Create Service
			service = r.constructService(kirillApp)
			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "Failed to create Service")
				return ctrl.Result{}, err
			}
		}
	} else {
		logger.Info("Service already exists", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
	}

	// Update the status of KirillApp
	if err := r.updateStatus(ctx, kirillApp, deployment); err != nil {
		logger.Error(err, "Failed to update KirillApp status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// constructDeployment constructs a Deployment object based on the KirillApp spec
func (r *KirillAppReconciler) constructDeployment(kirillApp *myappv1.KirillApp) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name,
			Namespace: kirillApp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &kirillApp.Spec.Replicas,
			Selector: kirillApp.Spec.Selector,
			Template: kirillApp.Spec.Template,
		},
	}
}

// constructService constructs a Service object based on the KirillApp spec
func (r *KirillAppReconciler) constructService(kirillApp *myappv1.KirillApp) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kirillApp.Name,
			Namespace: kirillApp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: kirillApp.Spec.Template.Labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

// deploymentEqual checks if two Deployment specs are equivalent
func deploymentEqual(existing *appsv1.Deployment, desired *appsv1.Deployment) bool {
	return existing.Spec.Replicas != desired.Spec.Replicas || existing.Spec.Template != desired.Spec.Template
}

// updateStatus updates the status of KirillApp based on the current Deployment state
func (r *KirillAppReconciler) updateStatus(ctx context.Context, kirillApp *myappv1.KirillApp, deployment *appsv1.Deployment) error {
	kirillApp.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	kirillApp.Status.ReadyReplicas = deployment.Status.ReadyReplicas

	return r.Status().Update(ctx, kirillApp)
}

func (r *KirillAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.KirillApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
