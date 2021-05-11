package cpython

// Cpython is the name of the layer into which cpython dependency is installed.
const Cpython = "cpython"

// DepKey is the key in the Layer Content Metadata used to determine if layer
// can be reused.
const DepKey = "dependency-sha"

// Priorities is a list of version-source values that may appear in
// the BuildpackPlan entries that the buildpack receives. The list is
// from highest priority to lowest priority and determines the precedence
// of version-sources.
var Priorities = []interface{}{
	"BP_CPYTHON_VERSION",
}
