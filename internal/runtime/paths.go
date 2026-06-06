package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	socketName = "statecast.sock"
	pidName    = "statecast.pid"
)

// Dir は statecast のランタイムファイル（ソケット・PID）を置くディレクトリを返す。
func Dir() (string, error) {
	if dir := os.Getenv("STATECAST_RUNTIME_DIR"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(home, ".local", "statecast"), nil
}

// SocketPath は Unix ドメインソケットのパスを返す。
func SocketPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, socketName), nil
}

// PIDPath は daemon PID ファイルのパスを返す。
func PIDPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, pidName), nil
}

// EnsureDir はランタイムディレクトリが存在することを保証する。
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create runtime directory: %w", err)
	}
	return dir, nil
}
