//go:build unix

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const startupWait = 5 * time.Second

// StartBackground は daemon をバックグラウンドプロセスとして起動する。
func StartBackground() error {
	if IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	cmd := exec.Command(executable, "start", "--foreground")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon process: %w", err)
	}

	deadline := time.Now().Add(startupWait)
	for time.Now().Before(deadline) {
		if IsRunning() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("daemon failed to start")
}
