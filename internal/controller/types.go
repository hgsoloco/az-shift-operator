package controller

import (
	zonalShiftTypes "github.com/hgsoloco/az-shift-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type AzShiftResult struct {
	AzShift         zonalShiftTypes.AvailabilityZoneShift
	Result          ctrl.Result
	ResourceUpdated bool
	Error           error
}

func (e *AzShiftResult) WasResourceUpdatedOrError() bool {
	return e.Error != nil || e.ResourceUpdated
}
