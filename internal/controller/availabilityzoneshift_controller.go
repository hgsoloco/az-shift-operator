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
	"fmt"
	zonalShiftTypes "github.com/hgsoloco/az-shift-operator/api/v1"
	//apiv1 "k8s.io/api/core/v1"
	v1alpha5 "github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AvailabilityZoneShiftReconciler reconciles a AvailabilityZoneShift object
type AvailabilityZoneShiftReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=zonalshiftmgr.keikoproj.io,resources=availabilityzoneshifts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zonalshiftmgr.keikoproj.io,resources=availabilityzoneshifts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=zonalshiftmgr.keikoproj.io,resources=availabilityzoneshifts/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=provisioners,scope=Cluster,categories=karpenter +kubebuilder:subresource:status

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AvailabilityZoneShift object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *AvailabilityZoneShiftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Fprintf(os.Stderr, "THIS IS RECCCCCCC")
	log := log.FromContext(ctx)
	log.Info("THIS IS LOG RECONCILE")

	azShift := zonalShiftTypes.AvailabilityZoneShift{Status: zonalShiftTypes.AvailabilityZoneShiftStatus{}}

	if err := r.Get(ctx, req.NamespacedName, &azShift); err != nil {
		log.Error(err, "unable to fetch az shift resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var prov v1alpha5.ProvisionerList
	var listOpts []client.ListOption

	if err := r.Client.List(ctx, &prov, listOpts...); err != nil {
		log.Error(err, "unable to fetch provisioner")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Provisioners Found", "Updating them..", len(prov.Items))

	impairedAz := azShift.Spec.ImpairedAZ
	log.Info("Impaired AZ: ", "Which is..", impairedAz)

	topologyKey := "topology.kubernetes.io/zone"

	var changedItems []int

	for index, value := range prov.Items {
		changed := false
		for k, v := range value.Spec.Requirements {
			if v.Key == topologyKey {
				changed = removeImpairedZone(&value.Spec.Requirements[k].Values, impairedAz)
			}
		}
		if changed == true {
			changedItems = append(changedItems, index)
		}
	}

	for _, m := range changedItems {
		if err := r.Update(ctx, &prov.Items[m]); err != nil {
			log.Error(err, "unable to update Provisioner status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func removeImpairedZone(values *[]string, impairedAz string) bool {
	changed := false
	x := *values
	for i, v := range x {
		if v == impairedAz {
			// Found the value, remove it by slicing
			*values = append(x[:i], x[i+1:]...)
			changed = true
			break
		}
	}
	return changed

	//for _, s := range values {
	//	if s == impairedAz {
	//		changedItems = append(changedItems, index)
	//	} else {
	//		for _, t := range goodZones {
	//			goodZones = append(goodZones, t)
	//			found := false
	//			if t == s {
	//				found = true
	//				break
	//			}
	//			if !found {
	//				goodZones = append(goodZones, t)
	//			}
	//		}
	//	}
	//}
	//value.Spec.Requirements[k].Values = goodZones
}

// SetupWithManager sets up the controller with the Manager.
func (r *AvailabilityZoneShiftReconciler) SetupWithManager(mgr ctrl.Manager) error {
	fmt.Fprintf(os.Stderr, "THIS IS SETUPWITHMANAGER")

	return ctrl.NewControllerManagedBy(mgr).
		For(&zonalShiftTypes.AvailabilityZoneShift{}).
		Complete(r)
}
