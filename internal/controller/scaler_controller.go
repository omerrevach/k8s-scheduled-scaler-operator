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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/omerrevach/k8s-scheduled-scaler-operator/api/v1alpha1"
)


// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.omerrevach.online,resources=scalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.omerrevach.online,resources=scalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.omerrevach.online,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// says that my controller is working fine
	log := log.FromContext(ctx)
	log.Info("Reconcile called")

	// creates an empty instance of my custom resource Scaler (defined in my CRD)
	scaler := &apiv1alpha1.Scaler{}
	err := r.Get(ctx, req.NamespacedName, scaler)
	if err != nil {
		return ctrl.Result{}, nil
	}
	location, err := time.LoadLocation(scaler.Spec.Timezone)
	if err != nil {
		log.Error(err, "Failed to load the timezone", "timezone: ", scaler.Spec.Timezone)
	}
	
	startTime, err := time.ParseInLocation("15:04", scaler.Spec.Start, location)
	if err != nil {
		log.Error(err, "Failed to parse the start time", "start", scaler.Spec.Start)
		return ctrl.Result{}, err
	}
	endTime, err := time.ParseInLocation("15:04", scaler.Spec.End, location)
	if err != nil {
		log.Error(err, "Failed to parse end time", "end", scaler.Spec.End)
		return ctrl.Result{}, err
	}

	currentTime := time.Now().In(location)
	var replicas int32 = scaler.Spec.Replicas

	log.Info("Start Time: ", "startTime", startTime)
	log.Info("End Time: ", "endTime", endTime)
	log.Info("Current Time: ", "currentTime", currentTime)

	if currentTime.After(startTime) && currentTime.Before(endTime) {
		for _, deploy := range scaler.Spec.Deployments {
			deployment := &appsv1.Deployment{}
			err := r.Get(ctx, types.NamespacedName{
				Namespace: deploy.Namespace, 
				Name: deploy.Name,
			},
				deployment,
			)
			if err != nil {
				log.Error(err, "Failed to get deployment", "name", deploy.Name, "namespace", deploy.Namespace)
				continue // Skip this deployment and move to the next one
			}

			if *deployment.Spec.Replicas != replicas {
				log.Info("Scaling Deployment", "name", deploy.Name, "namespace", deploy.Namespace)
				deployment.Spec.Replicas = &replicas
				err := r.Update(ctx, deployment)
				if err != nil {
					log.Error(err, "Failed to update deployment", "name", deploy.Name, "namespace", deploy.Namespace)
				}
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Scaler{}).
		Named("scaler").
		Complete(r)
}
