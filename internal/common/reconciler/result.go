package reconciler

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type ReconcileResult struct {
	Requeue      bool
	RequeueAfter time.Duration
}

func (r *ReconcileResult) RequeueOrNot() bool {
	if r.Requeue || r.RequeueAfter > 0 {
		return true
	}
	return false
}

func (r *ReconcileResult) Result() ctrl.Result {
	if r.RequeueOrNot() {
		return ctrl.Result{Requeue: r.Requeue, RequeueAfter: r.RequeueAfter}
	}
	return ctrl.Result{}
}
