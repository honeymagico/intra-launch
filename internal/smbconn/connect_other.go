//go:build !windows

package smbconn

// Connect is a no-op outside Windows, where UNC authentication is unavailable.
func Connect(remotePath, username, password string) error {
	return nil
}
