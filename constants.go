package cpython

// Cpython is the name of the layer into which cpython dependency is installed.
const Cpython = "cpython"

// TODO(restructure): Remove this after restructing is completed.
const Python = "python"

// DepKey is the key in the Layer Content Metadata used to determine if layer
// can be reused.
const DepKey = "dependency-sha"

var Priorities = []interface{}{"buildpack.yml"}
