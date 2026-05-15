package application

import "errors"

// ErrDependencyUnavailable reports that a required service dependency is absent.
var ErrDependencyUnavailable = errors.New("dependency unavailable")
