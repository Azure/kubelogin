//go:build unix

package token

import (
	"os"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

// acquireProcessLock attempts to acquire an exclusive file lock at the given path.
// It returns a function that releases the lock. If the lock cannot be acquired,
// it returns a no-op function so callers always get a valid unlock function.
func acquireProcessLock(path string) func() {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		klog.V(5).Infof("failed to open lock file %s: %v", path, err)
		return func() {}
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		klog.V(5).Infof("failed to acquire lock on %s: %v", path, err)
		f.Close()
		return func() {}
	}
	return func() {
		if err := unix.Flock(int(f.Fd()), unix.LOCK_UN); err != nil {
			klog.V(5).Infof("failed to release lock on %s: %v", path, err)
		}
		f.Close()
	}
}
