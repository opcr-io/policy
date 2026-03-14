package app //nolint: testpackage // importing private functions into test scope

// Make private functions accessible to the _test scope.

func IsPathSafe(absTarget, absAllowed string) bool {
	return isPathSafe(absTarget, absAllowed)
}
