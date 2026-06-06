package daemon

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/okonomipizza/statecast/internal/client"
)

func TestStopWhenNotRunning(t *testing.T) {
	// daemon 未起動時はエラーになること
	runtimeDir, err := os.MkdirTemp("/tmp", "sc")
	if err != nil {
		t.Fatalf("テスト用ランタイムディレクトリの作成に失敗した: %v", err)
	}
	t.Setenv("STATECAST_RUNTIME_DIR", runtimeDir)

	if err := Stop(); err == nil {
		t.Fatal("daemon 未起動時に Stop が成功してしまった")
	}
}

func TestStopStopsDaemonProcess(t *testing.T) {
	// stop が起動中の daemon プロセスを停止すること
	bin := buildTestBinary(t)

	runtimeDir, err := os.MkdirTemp("/tmp", "sc")
	if err != nil {
		t.Fatalf("テスト用ランタイムディレクトリの作成に失敗した: %v", err)
	}
	t.Setenv("STATECAST_RUNTIME_DIR", runtimeDir)

	daemonProc := exec.Command(bin, "start", "--foreground")
	daemonProc.Env = append(os.Environ(), "STATECAST_RUNTIME_DIR="+runtimeDir)
	if err := daemonProc.Start(); err != nil {
		t.Fatalf("daemon プロセスの起動に失敗した: %v", err)
	}
	t.Cleanup(func() {
		_ = daemonProc.Process.Kill()
		_ = daemonProc.Wait()
	})

	waitForDaemon(t)

	if err := Stop(); err != nil {
		t.Fatalf("stop に失敗した: %v", err)
	}

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}
	if err := c.Ping(); err == nil {
		t.Fatal("stop 後も daemon が応答している")
	}
}

func buildTestBinary(t *testing.T) string {
	t.Helper()

	root := moduleRoot(t)
	bin := filepath.Join(t.TempDir(), "statecast")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/statecast")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("テスト用バイナリのビルドに失敗した: %v\n%s", err, out)
	}
	return bin
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go.mod のパス取得に失敗した: %v", err)
	}

	modFile := strings.TrimSpace(string(out))
	if modFile == "" || modFile == "/dev/null" {
		t.Fatal("go module が見つからない")
	}

	return filepath.Dir(modFile)
}

func waitForDaemon(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		c, err := client.New()
		if err == nil && c.Ping() == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("daemon の起動待ちがタイムアウトした")
}
