package builder

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client
	Schema *runtime.Scheme
}
