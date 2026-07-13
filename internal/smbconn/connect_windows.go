//go:build windows

package smbconn

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	resourceTypeDisk     = 1
	resourceDisplayShare = 3
)

type netResource struct {
	scope        uint32
	resourceType uint32
	displayType  uint32
	usage        uint32
	localName    *uint16
	remoteName   *uint16
	comment      *uint16
	provider     *uint16
}

var wnetAddConnection2 = syscall.NewLazyDLL("mpr.dll").NewProc("WNetAddConnection2W")

// Connect authenticates the current Windows session to an SMB share.
// It does not disconnect an existing session, which may belong to the user.
func Connect(remotePath, username, password string) error {
	if username == "" {
		return nil
	}
	remoteName, err := syscall.UTF16PtrFromString(remotePath)
	if err != nil {
		return fmt.Errorf("SMB 路徑無效: %w", err)
	}
	user, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		return fmt.Errorf("SMB 帳號無效: %w", err)
	}
	secret, err := syscall.UTF16PtrFromString(password)
	if err != nil {
		return fmt.Errorf("SMB 密碼無效: %w", err)
	}
	resource := netResource{
		resourceType: resourceTypeDisk,
		displayType:  resourceDisplayShare,
		remoteName:   remoteName,
	}
	result, _, _ := wnetAddConnection2.Call(
		uintptr(unsafe.Pointer(&resource)),
		uintptr(unsafe.Pointer(secret)),
		uintptr(unsafe.Pointer(user)),
		0,
	)
	if result != 0 {
		return fmt.Errorf("Windows 錯誤 %d", result)
	}
	return nil
}
