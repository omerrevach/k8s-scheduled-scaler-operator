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
	// Logging to indicate reconciliation loop is running
	log := log.FromContext(ctx)
	log.Info("Reconcile called")

	// Fetch the Scaler custom resource
	scaler := &apiv1alpha1.Scaler{}
	err := r.Get(ctx, req.NamespacedName, scaler)
	if err != nil {
		log.Error(err, "Failed to fetch Scaler resource")
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not found errors
	}

	// Load the timezone specified in the Scaler CR
	location, err := time.LoadLocation(scaler.Spec.Timezone)
	if err != nil {
		log.Error(err, "Failed to load the timezone", "timezone", scaler.Spec.Timezone)
		return ctrl.Result{}, err
	}

	// Use today's date for correct time parsing
	currentDate := time.Now().Format("2006-01-02")
	startTime, err := time.ParseInLocation("2006-01-02 15:04", currentDate+" "+scaler.Spec.Start, location)
	if err != nil {
		log.Error(err, "Failed to parse the start time", "start", scaler.Spec.Start)
		return ctrl.Result{}, err
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04", currentDate+" "+scaler.Spec.End, location)
	if err != nil {
		log.Error(err, "Failed to parse end time", "end", scaler.Spec.End)
		return ctrl.Result{}, err
	}

	// Get the current time in the user's timezone
	currentTime := time.Now().In(location)

	// Debugging: Print parsed times
	log.Info(fmt.Sprintf("Parsed Times -> Start: %v | End: %v | Now: %v", startTime, endTime, currentTime))

	// Determine whether to scale up or down
	var desiredReplicas int32
	if currentTime.After(startTime) && currentTime.Before(endTime) {
		desiredReplicas = scaler.Spec.Replicas // Scale up
	} else {
		desiredReplicas = scaler.Spec.NormalReplicasAmount // Scale down
	}

	// Scale each deployment in the list
	for _, deploy := range scaler.Spec.Deployments {
		deployment := &appsv1.Deployment{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: deploy.Namespace,
			Name:      deploy.Name,
		}, deployment)
		if err != nil {
			log.Error(err, "Failed to get deployment", "name", deploy.Name, "namespace", deploy.Namespace)
			continue // Skip this deployment and move to the next one
		}

		// Ensure the replicas field is not nil before dereferencing
		if deployment.Spec.Replicas == nil {
			log.Info("Replicas field was nil, initializing to desired count", "name", deploy.Name, "namespace", deploy.Namespace, "replicas", desiredReplicas)
			deployment.Spec.Replicas = &desiredReplicas
		}

		// Only update if replicas need to be changed
		if *deployment.Spec.Replicas != desiredReplicas {
			log.Info("Scaling Deployment", "name", deploy.Name, "namespace", deploy.Namespace, "newReplicas", desiredReplicas)
			replicaCopy := desiredReplicas
			deployment.Spec.Replicas = &replicaCopy
			err := r.Update(ctx, deployment)
			if err != nil {
				log.Error(err, "Failed to update deployment", "name", deploy.Name, "namespace", deploy.Namespace)
			}
		}
	}

	// Requeue every 30 seconds to continuously check scaling conditions
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Scaler{}).
		Owns(&appsv1.Deployment{}). // Added to makes the controller watch Deployments
		Named("scaler").
		Complete(r)
}
