package operator

import (
	"context"

	"github.com/moodykhalif23/sme-platform/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MicroserviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Ensure MicroserviceReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &MicroserviceReconciler{}

func (r *MicroserviceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	var ms v1alpha1.Microservice
	if err := r.Client.Get(ctx, req.NamespacedName, &ms); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling Microservice", "namespace", ms.Namespace, "name", ms.Name)

	// Create or update Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.Name,
			Namespace: ms.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &ms.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": ms.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": ms.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  ms.Name,
						Image: ms.Spec.Image,
						Ports: []corev1.ContainerPort{{ContainerPort: ms.Spec.Port}},
					}},
				},
			},
		},
	}

	// Set the owner reference for garbage collection
	if err := controllerutil.SetControllerReference(&ms, deployment, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for deployment")
		return reconcile.Result{}, err
	}

	if err := r.Client.Create(ctx, deployment); err != nil {
		logger.Error(err, "Failed to create deployment")
		return reconcile.Result{}, err
	}

	// Create or update Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.Name,
			Namespace: ms.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": ms.Name,
			},
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       ms.Spec.Port,
				TargetPort: intstr.FromInt(int(ms.Spec.Port)),
			}},
		},
	}

	// Set the owner reference for garbage collection
	if err := controllerutil.SetControllerReference(&ms, service, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for service")
		return reconcile.Result{}, err
	}

	if err := r.Client.Create(ctx, service); err != nil {
		logger.Error(err, "Failed to create service")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Microservice{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(&MicroserviceReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		})
}
