package operator

import (
	"context"
	"fmt"

	"github.com/moodykhalif23/scalebit/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
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

func pointerToInt32(i int32) *int32 { return &i }

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
					Annotations: map[string]string{
						"linkerd.io/inject":    "enabled",
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   fmt.Sprintf("%d", ms.Spec.Port),
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

	// Create or update HorizontalPodAutoscaler
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.Name + "-hpa",
			Namespace: ms.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       ms.Name,
			},
			MinReplicas: pointerToInt32(1),
			MaxReplicas: 5,
			Metrics: []autoscalingv2.MetricSpec{{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: pointerToInt32(50),
					},
				},
			}},
		},
	}
	if err := controllerutil.SetControllerReference(&ms, hpa, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for HPA")
		return reconcile.Result{}, err
	}
	if err := r.Client.Create(ctx, hpa); err != nil {
		logger.Error(err, "Failed to create HPA")
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
