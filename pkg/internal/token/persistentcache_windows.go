//go:build windows

package token

// acquireProcessLock is a no-op on Windows because the macOS keychain
// race condition (issue #740) does not apply. Returns a no-op unlock function.
func acquireProcessLock(_ string) func() {
	return func() {}
}
