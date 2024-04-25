package master

import (
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
)

var _ builder.ContainerBuilder = &MastContainer{}

type MastContainer struct {
	builder.Container
}

func NewMastContainer() (*MastContainer, error) {
	return &MastContainer{}, nil
}
