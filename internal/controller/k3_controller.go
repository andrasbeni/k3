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

package controller

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webapplicationv1 "github.com/andrasbeni/k3/api/v1"
)

// K3Reconciler reconciles a K3 object
type K3Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapplication.abeni,resources=k3s,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapplication.abeni,resources=k3s/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapplication.abeni,resources=k3s/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *K3Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var instance = webapplicationv1.K3{}
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		logger.Error(err, "unable to fetch custom resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	replicas := int32(instance.Spec.Replicas)

	logger.Info("K3",
		"ReqNS", req.NamespacedName,
		"NS", instance.Namespace,
		"Image", instance.Spec.Image,
		"Replicas", instance.Spec.Replicas,
		"Host", instance.Spec.Host)

	if err := r.CreateOrUpdateDeployment(instance, replicas, ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.CreateOrUpdateService(instance, ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.CreateOrUpdateIngress(instance, ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *K3Reconciler) CreateOrUpdateDeployment(instance webapplicationv1.K3, replicas int32, ctx context.Context) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  instance.Name,
							Image: instance.Spec.Image,
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(&instance, deploy, r.Scheme); err != nil {
		return err
	}
	found := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found); err != nil && errors.IsNotFound(err) {
		err = r.Create(ctx, deploy)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}
	return nil
}

func (r *K3Reconciler) CreateOrUpdateService(instance webapplicationv1.K3, ctx context.Context) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-service",
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"deployment": instance.Name + "-deployment"},
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}
	if err := controllerutil.SetControllerReference(&instance, service, r.Scheme); err != nil {
		return err
	}
	found := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found); err != nil && errors.IsNotFound(err) {
		err = r.Create(ctx, service)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(service.Spec, found.Spec) {
		found.Spec = service.Spec
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}
	return nil
}

func (r *K3Reconciler) CreateOrUpdateIngress(instance webapplicationv1.K3, ctx context.Context) error {
	className := instance.Name + "-ingress-class"
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name + "-ingress",
			Namespace:   instance.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
				"cert-manager.io/issuer": instance.Name + "-issuer"},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &(className),
			TLS: []networkingv1.IngressTLS {{
				Hosts: []string {instance.Spec.Host},
				SecretName: instance.Name + "-tls",
			}},
			Rules: []networkingv1.IngressRule{{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: instance.Name + "-service",
									Port: networkingv1.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						}},
					},
				},
			}},
		},
	}

	if err := controllerutil.SetControllerReference(&instance, ingress, r.Scheme); err != nil {
		return err
	}
	found := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, found); err != nil && errors.IsNotFound(err) {
		err = r.Create(ctx, ingress)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(ingress.Spec, found.Spec) {
		found.Spec = ingress.Spec
		if err := r.Update(ctx, found); err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *K3Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webapplicationv1.K3{}).
		Complete(r)
}
