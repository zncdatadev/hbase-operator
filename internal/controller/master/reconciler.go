package master

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("master")
)
