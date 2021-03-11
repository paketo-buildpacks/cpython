package cpython

const (
	// Cpython is the name of the layer into which cpython dependency is
	// installed.
	Cpython = "cpython"

	// TODO(restructure): Remove this after restructing is completed.
	Python = "python"

	// DepKey is the key in the Layer Content Metadata used to determine if layer
	// can be reused.
	DepKey = "dependency-sha"
)
