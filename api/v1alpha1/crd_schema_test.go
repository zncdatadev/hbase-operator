package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zncdatadev/operator-go/pkg/testutil"
)

// TestNoInheritedConfigDefaults guards against the operator-go #544 defect class: a CRD
// `default` inside a role/role-group config block makes "unset here" indistinguishable from
// "explicitly the default", so a role group could never inherit the role's value. Defaults for
// those fields belong at consumption time, not in the schema.
func TestNoInheritedConfigDefaults(t *testing.T) {
	g := NewWithT(t)
	g.Expect("../../config/crd/bases/*.yaml").To(testutil.HaveNoInheritedConfigDefaults())
}
