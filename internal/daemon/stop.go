package daemon

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/okonomipizza/statecast/internal/client"
	"github.com/okonomipizza/statecast/internal/runtime"
)

const stopWait = 5 * time.Second

// Stop は起動中の daemon に SIGTERM を送り、停止するまで待つ。
func Stop() error {
	if !IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	if err := SignalProcess(syscall.SIGTERM); err != nil {
		if os.IsNotExist(err) {
			removeRuntimeFiles()
			return fmt.Errorf("daemon is not running")
		}
		return fmt.Errorf("stop daemon: %w", err)
	}

	deadline := time.Now().Add(stopWait)
	for time.Now().Before(deadline) {
		if !IsRunning() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("daemon failed to stop")
}

// IsRunning は daemon が応答可能かどうかを返す。
func IsRunning() bool {
	c, err := client.New()
	if err != nil {
		return false
	}
	return c.Ping() == nil
}

func removeRuntimeFiles() {
	socketPath, err := runtime.SocketPath()
	if err == nil {
		_ = os.Remove(socketPath)
	}
	pidPath, err := runtime.PIDPath()
	if err == nil {
		_ = os.Remove(pidPath)
	}
}
