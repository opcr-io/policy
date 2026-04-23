package app //nolint: testpackage // exporting private functions for test scope

// IsPathSafe exports isPathSafe for use in external test packages.
func IsPathSafe(absTarget, absAllowed string) bool {
	return isPathSafe(absTarget, absAllowed)
}
