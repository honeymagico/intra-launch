//go:build !windows

package launcher

// Running is only required on the target Windows platform.
func Running(string) (bool, error) { return false, nil }
